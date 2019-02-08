package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/techknowlogick/shiori/model"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gormigrate.v1"
)

// SQLiteDatabase is implementation of Database interface for connecting to SQLite3 database.
type GORMDatabase struct {
	*gorm.DB
}

// OpenSQLiteDatabase creates and open connection to new SQLite3 database.
func OpenGORMDatabase(dsn, dbType string) (*GormDatabase, error) {
	// Open database and start transaction
	db, err := gorm.Open(dbType, dsn)
	if err != nil {
		logrus.Fatalln(err)
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "initial",
			Migrate: func(tx *gorm.DB) error {
				// TODO: copy structs into here
				return tx.AutoMigrate(&model.Tag{}, &model.Bookmark{}, &model.Account{}, &model.LoginRequest{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return gormigrate.ErrRollbackImpossible
			},
		},
	})

	if err = m.Migrate(); err != nil {
		logrus.Fatalf("Could not migrate: %v", err)
	}

	return &GormDatabase{*db}, nil
}

// InsertBookmark inserts new bookmark to database. Returns new ID and error if any happened.
func (db *GormDatabase) InsertBookmark(bookmark model.Bookmark) (bookmarkID int, err error) {
	// Check URL and title
	if bookmark.URL == "" {
		return -1, fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return -1, fmt.Errorf("Title must not be empty")
	}

	// Set default ID and modified time
	if bookmark.ID == 0 {
		bookmark.ID, err = db.GetNewID("bookmark")
		if err != nil {
			return -1, err
		}
	}

	if bookmark.Modified == "" {
		bookmark.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return -1, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			bookmarkID = -1
			err = panicErr
		}
	}()

	// Save article to database
	tx.MustExec(`INSERT INTO bookmark (
		id, url, title, image_url, excerpt, author, 
		min_read_time, max_read_time, modified) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bookmark.ID,
		bookmark.URL,
		bookmark.Title,
		bookmark.ImageURL,
		bookmark.Excerpt,
		bookmark.Author,
		bookmark.MinReadTime,
		bookmark.MaxReadTime,
		bookmark.Modified)

	// Save bookmark content
	tx.MustExec(`INSERT INTO bookmark_content 
		(docid, title, content, html) VALUES (?, ?, ?, ?)`,
		bookmark.ID, bookmark.Title, bookmark.Content, bookmark.HTML)

	// Save tags
	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag (tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	for _, tag := range bookmark.Tags {
		tagName := strings.ToLower(tag.Name)
		tagName = strings.TrimSpace(tagName)

		tagID := -1
		err = stmtGetTag.Get(&tagID, tagName)
		checkError(err)

		if tagID == -1 {
			res := stmtInsertTag.MustExec(tagName)
			tagID64, err := res.LastInsertId()
			checkError(err)

			tagID = int(tagID64)
		}

		stmtInsertBookmarkTag.Exec(tagID, bookmark.ID)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	bookmarkID = bookmark.ID
	return bookmarkID, err
}

// GetBookmarks fetch list of bookmarks based on submitted ids.
func (db *GormDatabase) GetBookmarks(withContent bool, ids ...int) ([]model.Bookmark, error) {
	// Create query
	query := `SELECT 
		b.id, b.url, b.title, b.image_url, b.excerpt, b.author, 
		b.min_read_time, b.max_read_time, b.modified, bc.content <> "" has_content
		FROM bookmark b
		LEFT JOIN bookmark_content bc ON bc.docid = b.id`

	if withContent {
		query = `SELECT 
			b.id, b.url, b.title, b.image_url, b.excerpt, b.author, 
			b.min_read_time, b.max_read_time, b.modified, bc.content, bc.html, 
			bc.content <> "" has_content
			FROM bookmark b
			LEFT JOIN bookmark_content bc ON bc.docid = b.id`
	}

	// Prepare where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(ids) > 0 {
		whereClause = " WHERE b.id IN ("
		for _, id := range ids {
			args = append(args, id)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Fetch bookmarks
	query += whereClause
	bookmarks := []model.Bookmark{}
	err := db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
		FROM bookmark_tag bt LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ? ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer stmtGetTags.Close()

	for i, book := range bookmarks {
		book.Tags = []model.Tag{}
		err = stmtGetTags.Select(&book.Tags, book.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		bookmarks[i] = book
	}

	return bookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *GormDatabase) DeleteBookmarks(ids ...int) (err error) {
	// Create args and where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(ids) > 0 {
		whereClause = " WHERE id IN ("
		for _, id := range ids {
			args = append(args, id)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			err = panicErr
		}
	}()

	// Delete bookmarks
	whereTagClause := strings.Replace(whereClause, "id", "bookmark_id", 1)
	whereContentClause := strings.Replace(whereClause, "id", "docid", 1)

	tx.MustExec("DELETE FROM bookmark "+whereClause, args...)
	tx.MustExec("DELETE FROM bookmark_tag "+whereTagClause, args...)
	tx.MustExec("DELETE FROM bookmark_content "+whereContentClause, args...)

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return err
}

// SearchBookmarks search bookmarks by the keyword or tags.
func (db *GormDatabase) SearchBookmarks(orderLatest bool, keyword string, tags ...string) ([]model.Bookmark, error) {
	// Prepare query
	args := []interface{}{}
	query := `SELECT 
		b.id, b.url, b.title, b.image_url, b.excerpt, b.author, 
		b.min_read_time, b.max_read_time, b.modified, bc.content <> "" has_content
		FROM bookmark b
		LEFT JOIN bookmark_content bc ON bc.docid = b.id
		WHERE 1`

	// Create where clause for keyword
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		query += ` AND (b.url LIKE ? OR b.id IN (
			SELECT docid id FROM bookmark_content 
			WHERE title MATCH ? OR content MATCH ?))`
		args = append(args, "%"+keyword+"%", keyword, keyword)
	}

	// Create where clause for tags
	if len(tags) > 0 {
		whereTagClause := ` AND b.id IN (
			SELECT bookmark_id FROM bookmark_tag 
			WHERE tag_id IN (SELECT id FROM tag WHERE name IN (`

		for _, tag := range tags {
			args = append(args, tag)
			whereTagClause += "?,"
		}

		whereTagClause = whereTagClause[:len(whereTagClause)-1]
		whereTagClause += `)) GROUP BY bookmark_id HAVING COUNT(bookmark_id) >= ?)`
		args = append(args, len(tags))

		query += whereTagClause
	}

	// Set order clause
	if orderLatest {
		query += ` ORDER BY modified DESC`
	}

	// Fetch bookmarks
	bookmarks := []model.Bookmark{}
	err := db.Select(&bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Fetch tags for each bookmarks
	stmtGetTags, err := db.Preparex(`SELECT t.id, t.name 
		FROM bookmark_tag bt LEFT JOIN tag t ON bt.tag_id = t.id
		WHERE bt.bookmark_id = ? ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer stmtGetTags.Close()

	for i := range bookmarks {
		tags := []model.Tag{}
		err = stmtGetTags.Select(&tags, bookmarks[i].ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		bookmarks[i].Tags = tags
	}

	return bookmarks, nil
}

// UpdateBookmarks updates the saved bookmark in database.
func (db *GormDatabase) UpdateBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	// Prepare transaction
	tx, err := db.Beginx()
	if err != nil {
		return []model.Bookmark{}, err
	}

	// Make sure to rollback if panic ever happened
	defer func() {
		if r := recover(); r != nil {
			panicErr, _ := r.(error)
			tx.Rollback()

			result = []model.Bookmark{}
			err = panicErr
		}
	}()

	// Prepare statement
	stmtUpdateBookmark, err := tx.Preparex(`UPDATE bookmark SET
		url = ?, title = ?, image_url = ?, excerpt = ?, author = ?,
		min_read_time = ?, max_read_time = ?, modified = ? WHERE id = ?`)
	checkError(err)

	stmtUpdateBookmarkContent, err := tx.Preparex(`UPDATE bookmark_content SET
		title = ?, content = ?, html = ? WHERE docid = ?`)
	checkError(err)

	stmtGetTag, err := tx.Preparex(`SELECT id FROM tag WHERE name = ?`)
	checkError(err)

	stmtInsertTag, err := tx.Preparex(`INSERT INTO tag (name) VALUES (?)`)
	checkError(err)

	stmtInsertBookmarkTag, err := tx.Preparex(`INSERT OR IGNORE INTO bookmark_tag (tag_id, bookmark_id) VALUES (?, ?)`)
	checkError(err)

	stmtDeleteBookmarkTag, err := tx.Preparex(`DELETE FROM bookmark_tag WHERE bookmark_id = ? AND tag_id = ?`)
	checkError(err)

	result = []model.Bookmark{}
	for _, book := range bookmarks {
		// Save bookmark
		stmtUpdateBookmark.MustExec(
			book.URL,
			book.Title,
			book.ImageURL,
			book.Excerpt,
			book.Author,
			book.MinReadTime,
			book.MaxReadTime,
			book.Modified,
			book.ID)

		// Save bookmark content
		stmtUpdateBookmarkContent.MustExec(
			book.Title,
			book.Content,
			book.HTML,
			book.ID)

		// Save bookmark tags
		newTags := []model.Tag{}
		for _, tag := range book.Tags {
			if tag.Deleted {
				stmtDeleteBookmarkTag.MustExec(book.ID, tag.ID)
				continue
			}

			if tag.ID == 0 {
				tagID := -1
				err = stmtGetTag.Get(&tagID, tag.Name)
				checkError(err)

				if tagID == -1 {
					res := stmtInsertTag.MustExec(tag.Name)
					tagID64, err := res.LastInsertId()
					checkError(err)

					tagID = int(tagID64)
				}

				stmtInsertBookmarkTag.Exec(tagID, book.ID)
			}

			newTags = append(newTags, tag)
		}

		book.Tags = newTags
		result = append(result, book)
	}

	// Commit transaction
	err = tx.Commit()
	checkError(err)

	return result, err
}

// CreateAccount saves new account to database. Returns new ID and error if any happened.
func (db *GormDatabase) CreateAccount(username, password string) (err error) {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	// Insert account to database
	_, err = db.Exec(`INSERT INTO account
		(username, password) VALUES (?, ?)`,
		username, hashedPassword)

	return err
}

// GetAccount fetch account with matching username
func (db *GormDatabase) GetAccount(username string) (model.Account, error) {
	account := model.Account{}
	err := db.Get(&account,
		`SELECT id, username, password FROM account WHERE username = ?`,
		username)
	return account, err
}

// GetAccounts fetch list of accounts with matching keyword
func (db *GormDatabase) GetAccounts(keyword string) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	query := `SELECT id, username, password FROM account`

	if keyword == "" {
		query += " WHERE 1"
	} else {
		query += " WHERE username LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	query += ` ORDER BY username`

	// Fetch list account
	accounts := []model.Account{}
	err := db.Select(&accounts, query, args...)
	return accounts, err
}

// DeleteAccounts removes all record with matching usernames
func (db *GormDatabase) DeleteAccounts(usernames ...string) error {
	// Prepare where clause
	args := []interface{}{}
	whereClause := " WHERE 1"

	if len(usernames) > 0 {
		whereClause = " WHERE username IN ("
		for _, username := range usernames {
			args = append(args, username)
			whereClause += "?,"
		}

		whereClause = whereClause[:len(whereClause)-1]
		whereClause += ")"
	}

	// Delete usernames
	_, err := db.Exec(`DELETE FROM account `+whereClause, args...)
	return err
}

// GetTags fetch list of tags and their frequency
func (db *GormDatabase) GetTags() ([]model.Tag, error) {
	tags := []model.Tag{}
	query := `SELECT bt.tag_id id, t.name, COUNT(bt.tag_id) n_bookmarks 
		FROM bookmark_tag bt 
		LEFT JOIN tag t ON bt.tag_id = t.id
		GROUP BY bt.tag_id ORDER BY t.name`

	err := db.Select(&tags, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return tags, nil
}

// GetNewID creates new ID for specified table
func (db *GormDatabase) GetNewID(table string) (int, error) {
	var tableID int
	query := fmt.Sprintf(`SELECT IFNULL(MAX(id) + 1, 1) FROM %s`, table)

	err := db.Get(&tableID, query)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return tableID, nil
}

// GetBookmarkID fetchs bookmark ID based by its url
func (db *GormDatabase) GetBookmarkID(url string) int {
	var bookmarkID int
	db.Get(&bookmarkID, `SELECT id FROM bookmark WHERE url = ?`, url)
	return bookmarkID
}

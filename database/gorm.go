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
type GormDatabase struct {
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

	return &GormDatabase{db}, nil
}

// InsertBookmark inserts new bookmark to database. Returns new ID and error if any happened.
func (db *GormDatabase) InsertBookmark(bookmark model.Bookmark) (int, error) {
	// Check URL and title
	if bookmark.URL == "" {
		return -1, fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return -1, fmt.Errorf("Title must not be empty")
	}

	if bookmark.Modified == "" {
		bookmark.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return -1, err
	}
	if err := tx.Create(&bookmark).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return bookmark.ID, nil
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
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = r
		}
	}()
	if err := tx.Error; err != nil {
		return -1, err
	}

	if len(ids) > 0 {
		err = db.Where("id in (?)", ids).Delete(&model.Bookmark).Error
	} else {
		// if no IDs passed, then delete ALL book marks
		err = db.Delete(&model.Bookmark).Error
	}

	tx.Commit()

	return nil
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
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = r
			result = []model.Bookmark{}
		}
	}()
	if err := tx.Error; err != nil {
		return -1, err
	}

	result = []model.Bookmark{}
	for _, book := range bookmarks {
		if err := tx.Save(&book).Error; err != nil {
			tx.Rollback()
			return []model.Bookmark{}, err
		}
		result = append(result, book)
	}

	tx.Commit()

	return result, nil
}

// CreateAccount saves new account to database. Returns new ID and error if any happened.
func (db *GormDatabase) CreateAccount(username, password string) error {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}

	return db.Create(&model.Account{Username: username, Password: hashedPassword}).Error
}

// GetAccount fetch account with matching username
func (db *GormDatabase) GetAccount(username string) (model.Account, error) {
	account := model.Account{}
	err := db.Where("username = ?", username).First(&account).Error
	return account, err
}

// GetAccounts fetch list of accounts with matching keyword
func (db *GormDatabase) GetAccounts(keyword string) ([]model.Account, error) {
	accounts := []model.Account{}
	err := db.Where("username LIKE ?", "%"+keyword+"%").Find(&accounts).Error

	return accounts, err
}

// DeleteAccounts removes all record with matching usernames
func (db *GormDatabase) DeleteAccounts(usernames ...string) error {
	if len(usernames) > 0 {
		return db.Where("username in (?)", usernames).Delete(&model.Accounts).Error
	}
	// if no arg passed, then delete ALL accounts
	return db.Delete(&model.Accounts).Error
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

// GetBookmarkID fetchs bookmark ID based by its url
func (db *GormDatabase) GetBookmarkID(url string) int {
	bookmark := model.Bookmark{}
	db.Where("url = ?", url).First(&bookmark)
	return bookmark.ID
}

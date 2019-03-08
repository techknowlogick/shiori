package database

import (
	"fmt"
	"math"
	"strings"
	"time"

	"src.techknowlogick.com/shiori/model"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/builder"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteDatabase is implementation of Database interface for connecting to database.
type XormDatabase struct {
	*xorm.Engine
	dbType string
}

// OpenSQLiteDatabase creates and open connection to new database.
func OpenXormDatabase(dsn, dbType string) (*XormDatabase, error) {
	// Open database and start transaction
	db, err := xorm.NewEngine(dbType, dsn)
	if err != nil {
		return &XormDatabase{}, err
	}
	err = db.Sync2(new(model.Tag), new(model.Bookmark), new(model.BookmarkTag), new(model.Account))
	if err != nil {
		return &XormDatabase{}, err
	}
	return &XormDatabase{db, dbType}, nil
}

// InsertBookmark inserts new bookmark to database. Returns new ID and error if any happened.
func (db *XormDatabase) InsertBookmark(bookmark *model.Bookmark) error {
	// Check URL and title
	if bookmark.URL == "" {
		return fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return fmt.Errorf("Title must not be empty")
	}

	//	if bookmark.Modified == "" {
	bookmark.Modified = time.Now()
	//	}

	session := db.NewSession()
	defer session.Close()

	// add Begin() before any action
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return err
	}

	// create bookmark & get ID
	session.Insert(bookmark)
	for i := 0; i < len(bookmark.Tags); i++ {
		var tag model.Tag
		tag.Name = bookmark.Tags[i].Name
		has, err := session.Exist(&tag)
		if err != nil {
			return err
		}
		if !has {
			// create tag
			session.Insert(&tag)
		} else {
			session.Where("name = ?", tag.Name).Get(&tag)
		}
		bookmark.Tags[i] = tag
		// add bookmark_tag relation
		session.Insert(&model.BookmarkTag{BookmarkID: bookmark.ID, TagID: tag.ID})
	}
	session.Commit()
	return nil
}

// GetBookmarks fetch list of bookmarks based on submitted ids.
func (db *XormDatabase) GetBookmarks(withContent bool, ids ...int) ([]model.Bookmark, error) {
	bookmarks := make([]model.Bookmark, 0)
	var err error
	if len(ids) > 0 {
		err = db.In("id", ids).Find(&bookmarks)
	} else {
		err = db.Find(&bookmarks)
	}
	for i := 0; i < len(bookmarks); i++ {
		bookmarks[i].Tags = make([]model.Tag, 0)
		tags := make([]model.Tag, 0)
		bookmark := bookmarks[i]
		db.Join("left", "bookmark_tag", "bookmark_tag.tag_id = tag.id").Where(builder.Eq{"bookmark_tag.bookmark_id": bookmark.ID}).Find(&tags)
		bookmarks[i].Tags = tags
	}
	return bookmarks, err
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *XormDatabase) DeleteBookmarks(ids ...int) error {
	if len(ids) == 0 {
		return db.deleteBookmarks()
	}

	page := 0
	for len(ids) > page*100 {
		upperIndex := int(math.Min(float64(page*100+100), float64(len(ids))))
		err := db.deleteBookmarks(ids[page*100 : upperIndex]...)
		if err != nil {
			fmt.Println(err)
		}
		page = page + 1
	}
	return nil
}

// deleteBookmarks removes all record with matching ids from database
func (db *XormDatabase) deleteBookmarks(ids ...int) error {
	var bookmark model.Bookmark
	var err error
	if len(ids) > 0 {
		_, err = db.In("id", ids).Delete(&bookmark)
	} else {
		_, err = db.Delete(&bookmark)
	}
	return err
}

// SearchBookmarks search bookmarks by the keyword or tags.
func (db *XormDatabase) SearchBookmarks(orderLatest bool, keyword string, tags ...string) ([]model.Bookmark, error) {
	//var bookmarks []model.Bookmark
	bookmarks := make([]model.Bookmark, 0)
	searchCond := builder.NewCond()

	if len(keyword) > 0 {
		keyword = strings.TrimSpace(keyword)
		lowerKeyword := strings.ToLower(keyword)
		exprCond := builder.Or(
			builder.Like{"title", lowerKeyword},
			builder.Like{"content", lowerKeyword},
		)
		keywordCond := builder.Or(
			builder.Like{"url", lowerKeyword},
			exprCond,
		)
		searchCond = searchCond.And(keywordCond)
	}

	if len(tags) > 0 {
		tagsCond := builder.In("id", builder.Select("bookmark_id").From("bookmark_tag").LeftJoin("tag", builder.Expr("tag.id = bookmark_tag.tag_id")).Where(builder.In("tag.name", tags)))
		searchCond = searchCond.And(tagsCond)
	}

	err := db.Where(searchCond).Desc("created").Find(&bookmarks)

	for i := 0; i < len(bookmarks); i++ {
		bookmarks[i].Tags = make([]model.Tag, 0)
		tags := make([]model.Tag, 0)
		bookmark := bookmarks[i]
		db.Join("left", "bookmark_tag", "bookmark_tag.tag_id = tag.id").Where(builder.Eq{"bookmark_tag.bookmark_id": bookmark.ID}).Find(&tags)
		bookmarks[i].Tags = tags
	}

	return bookmarks, err
}

// UpdateBookmarks updates the saved bookmark in database.
func (db *XormDatabase) UpdateBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	result = []model.Bookmark{}
	session := db.NewSession()
	defer session.Close()

	// add Begin() before any action
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return []model.Bookmark{}, err
	}
	for _, bookmark := range bookmarks {
		// create bookmark & get ID
		session.Where("bookmark_id = ?", bookmark.ID).Update(&bookmark)
		// clear existing tag assignments
		session.Where("bookmark_id = ?", bookmark.ID).Delete(&model.BookmarkTag{})
		// insert & assign tag assignments
		for i := 0; i < len(bookmark.Tags); i++ {
			var tag model.Tag
			tag.Name = bookmark.Tags[i].Name
			has, err := session.Exist(&tag)
			if err != nil {
				return []model.Bookmark{}, err
			}
			if !has {
				// create tag
				session.Insert(&tag)
			} else {
				session.Where("name = ?", tag.Name).Get(&tag)
			}
			bookmark.Tags[i] = tag
			// add bookmark_tag relation
			session.Insert(&model.BookmarkTag{BookmarkID: bookmark.ID, TagID: tag.ID})
		}
		result = append(result, bookmark)
	}
	session.Commit()
	return result, nil
}

// CreateAccount saves new account to database. Returns new ID and error if any happened.
func (db *XormDatabase) CreateAccount(username, password string) error {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	_, err = db.Insert(&model.Account{Username: username, Password: string(hashedPassword)})
	return err
}

// GetAccount fetch account with matching username
func (db *XormDatabase) GetAccount(username string) (model.Account, error) {
	var account model.Account
	_, err := db.Where("username = ?", username).Get(&account)
	return account, err
}

// GetAccounts fetch list of accounts with matching keyword
func (db *XormDatabase) GetAccounts(keyword string) ([]model.Account, error) {
	var accounts []model.Account
	var err error
	if keyword == "" {
		err = db.Where(builder.Like{"username", keyword}).Find(&accounts)
	} else {
		err = db.Find(&accounts)
	}
	return accounts, err
}

// DeleteAccounts removes all record with matching usernames
func (db *XormDatabase) DeleteAccounts(usernames ...string) error {
	var account model.Account
	var err error
	if len(usernames) > 0 {
		_, err = db.Where("username in (?)", usernames).Delete(&account)
	} else {
		_, err = db.Delete(&account)
	}
	return err
}

// GetTags fetch list of tags and their frequency
func (db *XormDatabase) GetTags() ([]model.Tag, error) {
	tags := make([]model.Tag, 0)
	err := db.Table("tag").Select("bookmark_tag.tag_id as id, tag.name, COUNT(bookmark_tag.tag_id) as n_bookmarks").
		Join("left", "bookmark_tag", "bookmark_tag.tag_id = tag.id").
		GroupBy("bookmark_tag.tag_id, tag.name").Find(&tags)

	return tags, err
}

// GetBookmarkID fetchs bookmark ID based by its url
func (db *XormDatabase) GetBookmarkID(url string) int {
	var bookmark model.Bookmark
	db.Where("url = ?", url).Get(&bookmark)
	return bookmark.ID
}

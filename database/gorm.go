package database

import (
	"fmt"
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
func (db *GormDatabase) InsertBookmark(bookmark model.Bookmark) (uint, error) {
	// Check URL and title
	if bookmark.URL == "" {
		return uint(0), fmt.Errorf("URL must not be empty")
	}

	if bookmark.Title == "" {
		return uint(0), fmt.Errorf("Title must not be empty")
	}

	if bookmark.Modified == "" {
		bookmark.Modified = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			//return 0, r.(error)
		}
	}()
	if err := tx.Error; err != nil {
		return uint(0), err
	}
	if err := tx.Create(&bookmark).Error; err != nil {
		tx.Rollback()
		return uint(0), err
	}

	tx.Commit()

	return bookmark.ID, nil
}

// GetBookmarks fetch list of bookmarks based on submitted ids.
func (db *GormDatabase) GetBookmarks(withContent bool, ids ...uint) ([]model.Bookmark, error) {
	bookmarks := []model.Bookmark{}
	err := db.Find(&bookmarks).Error
	return bookmarks, err
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *GormDatabase) DeleteBookmarks(ids ...uint) (err error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = r.(error)
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}

	if len(ids) > 0 {
		err = db.Where("id in (?)", ids).Delete(&model.Bookmark{}).Error
	} else {
		// if no IDs passed, then delete ALL book marks
		err = db.Delete(&model.Bookmark{}).Error
	}

	tx.Commit()

	return nil
}

// SearchBookmarks search bookmarks by the keyword or tags.
func (db *GormDatabase) SearchBookmarks(orderLatest bool, keyword string, tags ...string) ([]model.Bookmark, error) {
	// TODO: complete this function
	return []model.Bookmark{}, nil
}

// UpdateBookmarks updates the saved bookmark in database.
func (db *GormDatabase) UpdateBookmarks(bookmarks ...model.Bookmark) (result []model.Bookmark, err error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = r.(error)
			result = []model.Bookmark{}
		}
	}()
	if err := tx.Error; err != nil {
		return []model.Bookmark{}, err
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

	return db.Create(&model.Account{Username: username, Password: string(hashedPassword)}).Error
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
		return db.Where("username in (?)", usernames).Delete(&model.Account{}).Error
	}
	// if no arg passed, then delete ALL accounts
	return db.Delete(&model.Account{}).Error
}

// GetTags fetch list of tags and their frequency
func (db *GormDatabase) GetTags() ([]model.Tag, error) {
	// TODO: complete this function
	return []model.Tag{}, nil
}

// GetBookmarkID fetchs bookmark ID based by its url
func (db *GormDatabase) GetBookmarkID(url string) uint {
	bookmark := model.Bookmark{}
	db.Where("url = ?", url).First(&bookmark)
	return bookmark.ID
}

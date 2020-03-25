package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"

	"xorm.io/xorm"
)

type M1Tag struct {
	ID        int `xorm:"'id' pk autoincr"`
	Name      string
	Deleted   bool
	NBookmark int       `xorm:"n_bookmarks"`
	Created   time.Time `xorm:"created"`
	Updated   time.Time `xorm:"updated"`
}

func (m M1Tag) TableName() string {
	return "tag"
}

type M1Bookmark struct {
	ID          int       `xorm:"'id' pk autoincr"`
	URL         string    `xorm:"url"`
	Title       string    `xorm:"'title' NOT NULL"`
	ImageURL    string    `xorm:"'image_url' NOT NULL"`
	Excerpt     string    `xorm:"'excerpt' NOT NULL"`
	Author      string    `xorm:"'author' NOT NULL"`
	MinReadTime int       `xorm:"'min_read_time' DEFAULT 0"`
	MaxReadTime int       `xorm:"'max_read_time' DEFAULT 0"`
	Modified    time.Time `xorm:"modified"`
	Content     string    `xorm:"TEXT 'content'"`
	HTML        string    `xorm:"TEXT 'html'"`
	HasContent  bool      `xorm:"has_content"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
}

func (m M1Bookmark) TableName() string {
	return "bookmark"
}

type M1BookmarkTag struct {
	BookmarkID int `xorm:"bookmark_id"`
	TagID      int `xorm:"tag_id"`
}

func (m M1BookmarkTag) TableName() string {
	return "bookmark_tag"
}

type M1Account struct {
	ID       int `xorm:"'id' pk autoincr"`
	Username string
	Password string
	Created  time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
}

func (m M1Account) TableName() string {
	return "account"
}

var (
	M1 = &xormigrate.Migration{
		ID:          "initial-migration",
		Description: "[M1] Create base set of tables",
		Migrate: func(tx *xorm.Engine) error {
			// Sync2 instead of CreateTables because tables may already exist
			return tx.Sync2(new(M1Tag), new(M1Bookmark), new(M1BookmarkTag), new(M1Account))
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(new(M1Tag), new(M1Bookmark), new(M1BookmarkTag), new(M1Account))
		},
	}
)

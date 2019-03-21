package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"

	"github.com/go-xorm/xorm"
)

type M1Tag struct {
	ID        int       `xorm:"'id' pk autoincr" json:"id"`
	Name      string    `json:"name"`
	Deleted   bool      `json:"-"`
	NBookmark int       `xorm:"n_bookmarks" json:"nBookmarks"`
	Created   time.Time `xorm:"created"`
	Updated   time.Time `xorm:"updated"`
}

func (m M1Tag) TableName() string {
	return "tag"
}

type M1Bookmark struct {
	ID          int       `xorm:"'id' pk autoincr" json:"id"`
	URL         string    `xorm:"url" json:"url"`
	Title       string    `xorm:"'title' NOT NULL" json:"title"`
	ImageURL    string    `xorm:"'image_url' NOT NULL" json:"imageURL"`
	Excerpt     string    `xorm:"'excerpt' NOT NULL" json:"excerpt"`
	Author      string    `xorm:"'author' NOT NULL" json:"author"`
	MinReadTime int       `xorm:"'min_read_time' DEFAULT 0"   json:"minReadTime"`
	MaxReadTime int       `xorm:"'max_read_time' DEFAULT 0"   json:"maxReadTime"`
	Modified    time.Time `xorm:"modified"    json:"modified"`
	Content     string    `xorm:"TEXT 'content'" json:"content"`
	HTML        string    `xorm:"TEXT 'html'" json:"html,omitempty"`
	HasContent  bool      `xorm:"has_content" json:"hasContent"`
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
	ID       int       `xorm:"'id' pk autoincr" json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
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
			return tx.Sync2(new(M1Tag), new(M1Bookmark), new(M1BookmarkTag), new(M1Account))
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(new(M1Tag), new(M1Bookmark), new(M1BookmarkTag), new(M1Account))
		},
	}
)

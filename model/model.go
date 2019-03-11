package model

import (
	//	"database/sql"
	"time"
)

// Tag is tag for the bookmark
type Tag struct {
	ID        int         `xorm:"'id' pk autoincr" json:"id"`
	Name      string      `json:"name"`
	Deleted   bool        `json:"-"`
	NBookmark int         `xorm:"n_bookmarks" json:"nBookmarks"`
	Bookmarks []*Bookmark `xorm:"-"`
	Created   time.Time   `xorm:"created"`
	Updated   time.Time   `xorm:"updated"`
}

// Bookmark is record of a specified URL
type Bookmark struct {
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
	Tags        []Tag     `xorm:"-"           json:"tags"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
}

type BookmarkTag struct {
	BookmarkID int `xorm:"bookmark_id"`
	TagID      int `xorm:"tag_id"`
}

// Account is account for accessing bookmarks from web interface
type Account struct {
	ID       int       `xorm:"'id' pk autoincr" json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
}

// LoginRequest is login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

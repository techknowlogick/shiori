package model

import (
	"github.com/jinzhu/gorm"
)

// Tag is tag for the bookmark
type Tag struct {
	gorm.Model
	Name       string `db:"name"        json:"name"`
	NBookmarks int    `db:"n_bookmarks" json:"nBookmarks"`
	Deleted    bool   `json:"-"`
}

// Bookmark is record of a specified URL
type Bookmark struct {
	gorm.Model
	URL         string `db:"url"           json:"url"`
	Title       string `db:"title"         json:"title"`
	ImageURL    string `db:"image_url"     json:"imageURL"`
	Excerpt     string `db:"excerpt"       json:"excerpt"`
	Author      string `db:"author"        json:"author"`
	MinReadTime int    `db:"min_read_time" json:"minReadTime"`
	MaxReadTime int    `db:"max_read_time" json:"maxReadTime"`
	Modified    string `db:"modified"      json:"modified"`
	Content     string `db:"content"       json:"-"`
	HTML        string `db:"html"          json:"html,omitempty"`
	HasContent  bool   `db:"has_content"   json:"hasContent"`
	Tags        []Tag  `json:"tags"`
}

// Account is account for accessing bookmarks from web interface
type Account struct {
	gorm.Model
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

// LoginRequest is login request
type LoginRequest struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

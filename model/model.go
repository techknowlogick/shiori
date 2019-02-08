package model

import (
	"github.com/jinzhu/gorm"
)

// Tag is tag for the bookmark
type Tag struct {
	gorm.Model
	Name      string `json:"name"`
	Deleted   bool
	Bookmarks []*Bookmark `gorm:"many2many:bookmark_tags;"`
}

// Bookmark is record of a specified URL
type Bookmark struct {
	gorm.Model
	URL         string `json:"url"`
	Title       string `json:"title"`
	ImageURL    string `json:"imageURL"`
	Excerpt     string `json:"excerpt"`
	Author      string `json:"author"`
	MinReadTime int    `json:"minReadTime"`
	MaxReadTime int    `json:"maxReadTime"`
	Modified    string `json:"modified"`
	Content     string `json:"-"`
	HTML        string `json:"html,omitempty"`
	HasContent  bool   `json:"hasContent"`
	Tags        []Tag  `gorm:"many2many:bookmark_tags;" json:"tags"`
}

// Account is account for accessing bookmarks from web interface
type Account struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest is login request
type LoginRequest struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

package cmd

import (
	"errors"
	"fmt"
	"math"
	nurl "net/url"
	"path/filepath"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-shiori/go-readability"
	"github.com/gofrs/uuid"
	"github.com/urfave/cli"
	"src.techknowlogick.com/shiori/model"
)

var (
	CmdAdd = cli.Command{
		Name:        "add",
		Description: "Bookmark the specified URL",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "title, i",
				Usage: "Custom title for this bookmark",
			},
			cli.StringFlag{
				Name:  "excerpt, e",
				Usage: "Custom excerpt for this bookmark",
			},
			cli.StringSliceFlag{
				Name:  "tags, t",
				Usage: "Comma-separated tags for this bookmark",
			},
			cli.BoolFlag{
				Name:  "offline, o",
				Usage: "Save bookmark without fetching data from internet",
			},
		},
		Action: runAddBookmark,
	}
)

func runAddBookmark(c *cli.Context) error {
	// Read flag and arguments
	args := c.Args()
	dataDir := c.GlobalString("data-dir")
	title := c.String("title")
	excerpt := c.String("excerpt")
	tags := c.StringSlice("tags")

	url := args[0]

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Make sure URL valid
	parsedURL, err := nurl.Parse(url)
	if err != nil || !valid.IsRequestURL(url) {
		return errors.New(cErrorSprint("URL is not valid"))
	}

	// Clear fragment and UTM parameters from URL
	parsedURL.Fragment = ""
	clearUTMParams(parsedURL)

	// Create bookmark item
	book := model.Bookmark{
		URL:     parsedURL.String(),
		Title:   normalizeSpace(title),
		Excerpt: normalizeSpace(excerpt),
	}

	// Set bookmark tags
	book.Tags = make([]model.Tag, len(tags))
	for i, tag := range tags {
		book.Tags[i].Name = strings.TrimSpace(tag)
	}

	// fetch data from internet
	article, _ := readability.FromURL(parsedURL.String(), 20*time.Second)

	book.Author = article.Byline
	book.MinReadTime = int(math.Floor(float64(article.Length)/(987+188) + 0.5))
	book.MaxReadTime = int(math.Floor(float64(article.Length)/(987-188) + 0.5))
	book.Content = article.TextContent
	book.HTML = article.Content

	// If title and excerpt doesnt have submitted value, use from article
	if book.Title == "" {
		book.Title = article.Title
	}

	if book.Excerpt == "" {
		book.Excerpt = strings.Map(fixUtf, article.Excerpt)
	}

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Save bookmark image to local disk
	u2, err := uuid.NewV4()
	if err != nil {
		return errors.New(cErrorSprint(err))
	}
	imgPath := filepath.Join(dataDir, "thumb", u2.String())
	err = downloadFile(article.Image, imgPath, 20*time.Second)
	if err == nil {
		book.ImageURL = fmt.Sprintf("/thumb/%s", u2)
	}

	// Save bookmark to database
	err = db.InsertBookmark(&book)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	printBookmarks(book)

	return nil
}

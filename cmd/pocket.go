package cmd

import (
	"errors"
	nurl "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"src.techknowlogick.com/shiori/model"

	"github.com/PuerkitoBio/goquery"
	valid "github.com/asaskevich/govalidator"
	"github.com/urfave/cli"
)

var (
	CmdPocket = cli.Command{
		Name:   "pocket",
		Usage:  "Import bookmarks from Pocket's exported HTML file",
		Action: runImportPocket,
	}
)

func runImportPocket(c *cli.Context) error {
	args := c.Args()

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	if len(args) != 1 {
		return errors.New(cErrorSprint("Please set path to source-file"))
	}

	// Open bookmark's file
	srcFile, err := os.Open(args[0])
	if err != nil {
		return errors.New(cErrorSprint(err))
	}
	defer srcFile.Close()

	// Parse bookmark's file
	doc, err := goquery.NewDocumentFromReader(srcFile)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	bookmarks := []model.Bookmark{}
	doc.Find("a").Each(func(_ int, a *goquery.Selection) {
		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")
		strModified, _ := a.Attr("time_added")
		intModified, _ := strconv.ParseInt(strModified, 10, 64)
		modified := time.Unix(intModified, 0)

		// Make sure URL valid
		parsedURL, err := nurl.Parse(url)
		if err != nil || !valid.IsRequestURL(url) {
			cError.Printf("%s will be skipped: URL is not valid\n\n", url)
			return
		}

		// Clear fragment and UTM parameters from URL
		parsedURL.Fragment = ""
		clearUTMParams(parsedURL)

		// Get bookmark tags
		tags := []model.Tag{}
		for _, strTag := range strings.Split(strTags, ",") {
			if strTag != "" {
				tags = append(tags, model.Tag{Name: strTag})
			}
		}

		// Add item to list
		bookmark := model.Bookmark{
			URL:      parsedURL.String(),
			Title:    normalizeSpace(title),
			Modified: modified,
			Tags:     tags,
		}

		bookmarks = append(bookmarks, bookmark)
	})

	// Save bookmarks to database
	for _, book := range bookmarks {
		// Save book to database
		err = db.InsertBookmark(&book)
		if err != nil {
			return errors.New(cErrorSprint("%s is skipped: %v\n\n", book.URL, err))
		}

		printBookmarks(book)
	}
	return nil
}

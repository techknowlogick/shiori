package cmd

import (
	"errors"
	"fmt"
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
	CmdImport = cli.Command{
		Name:        "import",
		Usage:       "import source-file",
		Description: "Import bookmarks from HTML file in Netscape Bookmark format",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "generate-tag, t",
				Usage: "Auto generate tag from bookmark's category",
			},
		},
	}
)

func runImportBookmarks(c *cli.Context) error {
	// Parse flags
	generateTag := c.Bool("generate-tag")
	args := c.Args()

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// If user doesn't specify, ask if tag need to be generated
	if !generateTag {
		var submit string
		fmt.Print("Add parents folder as tag? (y/n): ")
		fmt.Scanln(&submit)

		generateTag = submit == "y"
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
	doc.Find("dt>a").Each(func(_ int, a *goquery.Selection) {
		// Get related elements
		dt := a.Parent()
		dl := dt.Parent()

		// Get metadata
		title := a.Text()
		url, _ := a.Attr("href")
		strTags, _ := a.Attr("tags")
		strModified, _ := a.Attr("last_modified")
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

		// Get bookmark excerpt
		excerpt := ""
		if dd := dt.Next(); dd.Is("dd") {
			excerpt = dd.Text()
		}

		// Get category name for this bookmark
		// and add it as tags (if necessary)
		category := ""
		if dtCategory := dl.Prev(); dtCategory.Is("h3") {
			category = dtCategory.Text()
			category = normalizeSpace(category)
			category = strings.ToLower(category)
			category = strings.Replace(category, " ", "-", -1)
		}

		if category != "" && generateTag {
			tags = append(tags, model.Tag{Name: category})
		}

		// Add item to list
		bookmark := model.Bookmark{
			URL:      parsedURL.String(),
			Title:    normalizeSpace(title),
			Excerpt:  normalizeSpace(excerpt),
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
			return errors.New(cErrorSprint(fmt.Sprintf("%s is skipped: %v\n\n", book.URL, err)))
		}

		printBookmarks(book)
	}

	return nil
}

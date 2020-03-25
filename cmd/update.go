package cmd

import (
	"errors"
	"fmt"
	"math"
	nurl "net/url"
	fp "path/filepath"
	"strings"
	"sync"
	"time"

	"src.techknowlogick.com/shiori/database"
	"src.techknowlogick.com/shiori/utils"

	valid "github.com/asaskevich/govalidator"
	"github.com/go-shiori/go-readability"
	"github.com/gofrs/uuid"
	"github.com/gosuri/uiprogress"
	"github.com/urfave/cli"
	"src.techknowlogick.com/shiori/model"
)

var (
	CmdUpdate = cli.Command{
		Name:  "update",
		Usage: "Update the saved bookmarks",
		Description: "Update fields of an existing bookmark. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be updated. Update works differently depending on the flags:\n" +
			"- If indices are passed without any flags (--url, --title, --tag and --excerpt), read the URLs from DB and update titles from web.\n" +
			"- If --url is passed (and --title is omitted), update the title from web using the URL. While using this flag, update only accept EXACTLY one index.\n" +
			"While updating bookmark's tags, you can use - to remove tag (e.g. -nature to remove nature tag from this bookmark).",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "url, u",
				Usage: "New URL for this bookmark",
			},
			cli.StringFlag{
				Name:  "title, i",
				Usage: "New title for this bookmark",
			},
			cli.StringFlag{
				Name:  "excerpt, e",
				Usage: "New excerpt for this bookmark",
			},
			cli.StringSliceFlag{
				Name:  "tags, t",
				Usage: "Comma-separated tags for this bookmark",
			},
			cli.BoolFlag{
				Name:  "offline, o",
				Usage: "Update bookmark without fetching data from internet",
			},
			cli.BoolFlag{
				Name:  "yes, y",
				Usage: "Skip confirmation prompt and update ALL bookmarks",
			},
			cli.BoolFlag{
				Name:  "dont-overwrite",
				Usage: "Don't overwrite existing metadata. Useful when only want to update bookmark's content",
			},
		},
		Action: runUpdateBookmarks,
	}
)

func runUpdateBookmarks(c *cli.Context) error {
	// Parse flags
	args := c.Args()
	dataDir := c.GlobalString("data-dir")
	url := c.String("url")
	title := c.String("title")
	excerpt := c.String("excerpt")
	tags := c.StringSlice("tags")
	offline := c.Bool("offline")
	skipConfirm := c.Bool("yes")
	dontOverwrite := c.Bool("dont-overwrite")

	title = normalizeSpace(title)
	excerpt = normalizeSpace(excerpt)

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// Convert args to ids
	ids, err := utils.ParseIndexList(args)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// Check if --url flag is used
	if c.IsSet("url") {
		// Make sure URL is valid
		parsedURL, err := nurl.Parse(url)
		if err != nil || !valid.IsRequestURL(url) {
			return errors.New(utils.CErrorSprint("URL is not valid"))
		}

		// Clear fragment and UTM parameters from URL
		parsedURL.Fragment = ""
		utils.ClearUTMParams(parsedURL)
		url = parsedURL.String()

		// Make sure there is only one arguments
		if len(ids) != 1 {
			return errors.New(utils.CErrorSprint("Update only accepts one index while using --url flag"))
		}
	}

	// If no arguments (i.e all bookmarks will be updated),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmUpdate := ""
		fmt.Print("Update ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmUpdate)

		if confirmUpdate != "y" {
			return errors.New(utils.CErrorSprint("No bookmarks updated"))
		}
	}

	// Prepare wait group and mutex
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Fetch bookmarks from database
	bookmarks, err := db.GetBookmarks(database.BookmarkOptions{}, ids...)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	if len(bookmarks) == 0 {
		return errors.New(utils.CErrorSprint("No matching index found"))
	}

	// If not offline, fetch articles from internet
	listErrorMsg := []string{}
	if !offline {
		fmt.Println("Fetching new bookmarks data")

		// Start progress bar
		uiprogress.Start()
		bar := uiprogress.AddBar(len(bookmarks)).AppendCompleted().PrependElapsed()

		for i, book := range bookmarks {
			wg.Add(1)

			go func(pos int, book model.Bookmark) {
				// Make sure to increase bar
				defer func() {
					bar.Incr()
					wg.Done()
				}()

				// If used, use submitted URL
				if url != "" {
					book.URL = url
				}

				// Parse URL
				parsedURL, err := nurl.Parse(book.URL)
				if err != nil || !valid.IsRequestURL(book.URL) {
					mx.Lock()
					errorMsg := fmt.Sprintf("Failed to fetch %s: URL is not valid", book.URL)
					listErrorMsg = append(listErrorMsg, errorMsg)
					mx.Unlock()
					return
				}

				// Fetch data from internet
				article, err := readability.FromURL(parsedURL.String(), 20*time.Second)
				if err != nil {
					mx.Lock()
					errorMsg := fmt.Sprintf("Failed to fetch %s: %v", book.URL, err)
					listErrorMsg = append(listErrorMsg, errorMsg)
					mx.Unlock()
					return
				}

				book.Author = article.Byline
				book.MinReadTime = int(math.Floor(float64(article.Length)/(987+188) + 0.5))
				book.MaxReadTime = int(math.Floor(float64(article.Length)/(987-188) + 0.5))
				book.Content = article.TextContent
				book.HTML = article.Content

				if !dontOverwrite {
					book.Title = article.Title
					book.Excerpt = article.Excerpt
				}

				// Save bookmark image to local disk
				u2, err := uuid.NewV4()
				if err != nil {
					mx.Lock()
					errorMsg := fmt.Sprintf("Failed generate uuid")
					listErrorMsg = append(listErrorMsg, errorMsg)
					mx.Unlock()
					return
				}
				imgPath := fp.Join(dataDir, "thumb", u2.String())
				err = downloadFile(article.Image, imgPath, 20*time.Second)
				if err == nil {
					book.ImageURL = fmt.Sprintf("/thumb/%s", u2)
				}

				// Update list of bookmarks
				mx.Lock()
				bookmarks[pos] = book
				mx.Unlock()
			}(i, book)
		}

		wg.Wait()
		uiprogress.Stop()

		// Print error message
		fmt.Println()
		for _, errorMsg := range listErrorMsg {
			utils.CError.Println(errorMsg + "\n")
		}
	}

	// Map the tags to be added or deleted from flag --tags
	addedTags := make(map[string]struct{})
	deletedTags := make(map[string]struct{})
	for _, tag := range tags {
		tagName := strings.ToLower(tag)
		tagName = strings.TrimSpace(tagName)

		if strings.HasPrefix(tagName, "-") {
			tagName = strings.TrimPrefix(tagName, "-")
			deletedTags[tagName] = struct{}{}
		} else {
			addedTags[tagName] = struct{}{}
		}
	}

	// Set title, excerpt and tags from user submitted value
	for i, bookmark := range bookmarks {
		// Check if user submit his own title or excerpt
		if title != "" {
			bookmark.Title = title
		}

		if excerpt != "" {
			bookmark.Excerpt = excerpt
		}

		// Make sure title is not empty
		if bookmark.Title == "" {
			bookmark.Title = bookmark.URL
		}

		// Generate new tags
		tempAddedTags := make(map[string]struct{})
		for key, value := range addedTags {
			tempAddedTags[key] = value
		}

		newTags := []model.Tag{}
		for _, tag := range bookmark.Tags {
			if _, isDeleted := deletedTags[tag.Name]; isDeleted {
				tag.Deleted = true
			}

			if _, alreadyExist := addedTags[tag.Name]; alreadyExist {
				delete(tempAddedTags, tag.Name)
			}

			newTags = append(newTags, tag)
		}

		for tag := range tempAddedTags {
			newTags = append(newTags, model.Tag{Name: tag})
		}

		bookmark.Tags = newTags

		// Set bookmark new data
		bookmarks[i] = bookmark
	}

	// Update database
	result, err := db.UpdateBookmarks(bookmarks...)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// Print update result
	printBookmarks(result...)
	return nil
}

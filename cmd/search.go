package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

var (
	CmdSearch = cli.Command{
		Name:  "search",
		Usage: "Search bookmarks by submitted keyword",
		Description: "Search bookmarks by looking for matching keyword in bookmark's title and content. " +
			"If no keyword submitted, print all saved bookmarks. ",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "index-only, i",
				Usage: "Only print the index of bookmarks",
			},
			cli.BoolFlag{
				Name:  "json, j",
				Usage: "Output data in JSON format",
			},
			cli.StringSliceFlag{
				Name:  "tags, t",
				Usage: "Search bookmarks with specified tag(s)",
			},
		},
		Action: runSearchBookmarks,
	}
)

func runSearchBookmarks(c *cli.Context) error {
	// Read flags
	tags := c.StringSlice("tags")
	useJSON := c.Bool("json")
	indexOnly := c.Bool("index-only")
	args := c.Args()

	// Fetch keyword
	keyword := ""
	if len(args) > 0 {
		keyword = args[0]
	}

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Read bookmarks from database
	bookmarks, err := db.SearchBookmarks(false, keyword, tags...)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	if len(bookmarks) == 0 {
		return errors.New(cErrorSprint("No matching bookmarks found"))
	}

	// Print data
	if useJSON {
		bt, err := json.MarshalIndent(&bookmarks, "", "    ")
		if err != nil {
			return errors.New(cErrorSprint(err))
		}

		fmt.Println(string(bt))
		return nil
	}

	if indexOnly {
		for _, bookmark := range bookmarks {
			fmt.Printf("%d ", bookmark.ID)
		}
		fmt.Println()
		return nil
	}

	printBookmarks(bookmarks...)
	return nil
}

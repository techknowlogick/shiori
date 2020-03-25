package cmd

import (
	"errors"
	"fmt"
	"strings"

	"src.techknowlogick.com/shiori/database"
	"src.techknowlogick.com/shiori/utils"

	"github.com/urfave/cli"
)

var (
	CmdOpen = cli.Command{
		Name:  "open",
		Usage: "Open the saved bookmarks",
		Description: "Open bookmarks in browser. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, ALL bookmarks will be opened.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "yes, y",
				Usage: "Skip confirmation prompt and open ALL bookmarks",
			},
			cli.BoolFlag{
				Name:  "cache, c",
				Usage: "Open the bookmark's cache in text-only mode",
			},
			cli.BoolFlag{
				Name:  "trim-space",
				Usage: "Trim all spaces and newlines from the bookmark's cache",
			},
		},
		Action: runOpenBookmark,
	}
)

func runOpenBookmark(c *cli.Context) error {
	cacheMode := c.Bool("cache")
	trimSpace := c.Bool("trim-space")
	skipConfirm := c.Bool("yes")
	args := c.Args()

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// If no arguments (i.e all bookmarks will be opened),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmOpen := ""
		fmt.Print("Open ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmOpen)

		if confirmOpen != "y" {
			return nil
		}
	}

	// Convert args to ids
	ids, err := utils.ParseIndexList(args)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}
	bookmarks, err := db.GetBookmarks(database.BookmarkOptions{}, ids...)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	if len(bookmarks) == 0 {
		if len(args) > 0 {
			return errors.New(utils.CErrorSprint("No matching index found"))
		} else {
			return errors.New(utils.CErrorSprint("No saved bookmarks yet"))
		}
	}

	// If not cache mode, open bookmarks in browser
	if !cacheMode {
		for _, book := range bookmarks {
			err = openBrowser(book.URL)
			if err != nil {
				return errors.New(utils.CErrorSprint("Failed to open %s: %v\n", book.URL, err))
			}
		}
		return nil
	}

	termWidth := getTerminalWidth()
	if termWidth < 50 {
		termWidth = 50
	}

	for _, book := range bookmarks {
		if trimSpace {
			words := strings.Fields(book.Content)
			book.Content = strings.Join(words, " ")
		}

		utils.CIndex.Printf("%d. ", book.ID)
		utils.CTitle.Println(book.Title)
		fmt.Println()

		if book.Content == "" {
			utils.CError.Println("This bookmark doesn't have any cached content")
		} else {
			fmt.Println(book.Content)
		}

		fmt.Println()
		utils.CSymbol.Println(strings.Repeat("-", termWidth))
		fmt.Println()
	}

	return nil
}

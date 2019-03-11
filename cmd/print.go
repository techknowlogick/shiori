package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

var (
	CmdPrint = cli.Command{
		Name:    "print",
		Usage:   "print [indices]",
		Aliases: []string{"list", "ls"},
		Description: "Show the saved bookmarks by its DB index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records with actual index from DB are shown.",
		// hdl.printBookmarks
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "index-only, i",
				Usage: "Only print the index of bookmarks",
			},
			cli.BoolFlag{
				Name:  "json, j",
				Usage: "Output data in JSON format",
			},
		},
		Action: runPrintBookmarks,
	}
)

func runPrintBookmarks(c *cli.Context) error {
	// Read flags
	useJSON := c.Bool("json")
	indexOnly := c.Bool("index-only")
	args := c.Args()

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Read bookmarks from database
	bookmarks, err := db.GetBookmarks(false, ids...)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	if len(bookmarks) == 0 {
		if len(args) > 0 {
			return errors.New(cErrorSprint("No matching index found"))
		} else {
			return errors.New(cErrorSprint("No bookmarks saved yet"))
		}
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

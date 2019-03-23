package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"src.techknowlogick.com/shiori/utils"

	"github.com/urfave/cli"
)

var (
	CmdDelete = cli.Command{
		Name:  "delete",
		Usage: "Delete the saved bookmarks",
		Description: "Delete bookmarks. " +
			"When a record is deleted, the last record is moved to the removed index. " +
			"Accepts space-separated list of indices (e.g. 5 6 23 4 110 45), hyphenated range (e.g. 100-200) or both (e.g. 1-3 7 9). " +
			"If no arguments, all records will be deleted.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "yes, y",
				Usage: "Skip confirmation prompt and delete ALL bookmarks",
			},
		},
		Action: runDeleteBookmark,
	}
)

func runDeleteBookmark(c *cli.Context) error {
	// Read flag and arguments
	args := c.Args()
	dataDir := c.GlobalString("data-dir")
	skipConfirm := c.Bool("yes")

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// If no arguments (i.e all bookmarks going to be deleted),
	// confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmDelete := ""
		fmt.Print("Remove ALL bookmarks? (y/n): ")
		fmt.Scanln(&confirmDelete)
		if confirmDelete != "y" {
			return errors.New(utils.CErrorSprint("No bookmarks deleted"))
		}
	}

	// Convert args to ids
	ids, err := parseIndexList(args)
	if err != nil {
		return errors.New(utils.CErrorSprint(err))
	}

	// Delete bookmarks from database
	err = db.DeleteBookmarks(ids...)
	if err != nil {
		utils.CError.Println(err)
	}

	// Delete thumbnail image from local disk
	for _, id := range ids {
		// TODO: this logic is broken due to bookmark images using UUIDs
		imgPath := filepath.Join(dataDir, "thumb", fmt.Sprintf("%d", id))
		os.Remove(imgPath)
	}

	fmt.Println("Bookmark(s) have been deleted")
	return nil
}

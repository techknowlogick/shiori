package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	fp "path/filepath"
	"runtime"
	"strings"
	"time"

	"src.techknowlogick.com/shiori/database"
	"src.techknowlogick.com/shiori/model"
	"src.techknowlogick.com/shiori/utils"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func normalizeSpace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

func downloadFile(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure destination directory exist
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return err
	}

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write response body to the file
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// openBrowser tries to open the URL in a browser,
// and returns whether it succeed in doing so.
func openBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Run()
}

func getTerminalWidth() int {
	width, _, _ := terminal.GetSize(int(os.Stdin.Fd()))
	return width
}

func getDbConnection(c *cli.Context) (database.Database, error) {
	dbType := c.GlobalString("db-type")
	dbDsn := c.GlobalString("db-dsn")
	dataDir := c.GlobalString("data-dir")

	if dbType == "sqlite3" && dbDsn == "shiori.db" {
		dbDsn = filepath.Join(dataDir, dbDsn)
	}

	db, err := database.OpenXormDatabase(&database.XormOptions{DbDsn: dbDsn, DbType: dbType, ShowSQL: c.GlobalBool("show-sql-log")})
	return db, err

}

func printBookmarks(bookmarks ...model.Bookmark) {
	for _, bookmark := range bookmarks {
		// Create bookmark index
		strBookmarkIndex := fmt.Sprintf("%d. ", bookmark.ID)
		strSpace := strings.Repeat(" ", len(strBookmarkIndex))

		// Print bookmark title
		utils.CIndex.Print(strBookmarkIndex)
		utils.CTitle.Print(bookmark.Title)

		// Print read time
		if bookmark.MinReadTime > 0 {
			readTime := fmt.Sprintf(" (%d-%d minutes)", bookmark.MinReadTime, bookmark.MaxReadTime)
			if bookmark.MinReadTime == bookmark.MaxReadTime {
				readTime = fmt.Sprintf(" (%d minutes)", bookmark.MinReadTime)
			}
			utils.CReadTime.Println(readTime)
		} else {
			fmt.Println()
		}

		// Print bookmark URL
		utils.CSymbol.Print(strSpace + "> ")
		utils.CURL.Println(bookmark.URL)

		// Print bookmark excerpt
		if bookmark.Excerpt != "" {
			utils.CSymbol.Print(strSpace + "+ ")
			utils.CExcerpt.Println(bookmark.Excerpt)
		}

		// Print bookmark tags
		if len(bookmark.Tags) > 0 {
			utils.CSymbol.Print(strSpace + "# ")
			for i, tag := range bookmark.Tags {
				if i == len(bookmark.Tags)-1 {
					utils.CTag.Println(tag.Name)
				} else {
					utils.CTag.Print(tag.Name + ", ")
				}
			}
		}

		// Append new line
		fmt.Println()
	}
}

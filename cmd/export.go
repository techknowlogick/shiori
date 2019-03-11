package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"src.techknowlogick.com/shiori/model"

	"github.com/urfave/cli"
)

var (
	CmdExport = cli.Command{
		Name:   "export",
		Usage:  "Export bookmarks into HTML file in Netscape Bookmark format",
		Action: runExportBookmarks,
	}
)

func runExportBookmarks(c *cli.Context) error {
	args := c.Args()

	if len(args) != 1 {
		return errors.New(cErrorSprint("Please set path to target-file"))
	}

	db, err := getDbConnection(c)

	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Fetch bookmarks from database
	bookmarks, err := db.GetBookmarks(false)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	if len(bookmarks) == 0 {
		return errors.New(cErrorSprint("No saved bookmarks yet"))
	}

	// Make sure destination directory exist
	dstDir := filepath.Dir(args[0])
	os.MkdirAll(dstDir, os.ModePerm)

	// Open destination file
	dstFile, err := os.Create(args[0])
	if err != nil {
		return errors.New(cErrorSprint(err))
	}
	defer dstFile.Close()

	// Create template
	funcMap := template.FuncMap{
		"unix": func(t time.Time) int64 {
			return t.Unix()
		},
		"combine": func(tags []model.Tag) string {
			strTags := make([]string, len(tags))
			for i, tag := range tags {
				strTags[i] = tag.Name
			}

			return strings.Join(strTags, ",")
		},
	}

	tplContent := `<!DOCTYPE NETSCAPE-Bookmark-file-1>` +
		`<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">` +
		`<TITLE>Bookmarks</TITLE>` +
		`<H1>Bookmarks</H1>` +
		`<DL><p>` +
		`{{range $book := .}}` +
		`<DT><A HREF="{{$book.URL}}" ADD_DATE="{{$book.Modified | unix}}" TAGS="{{combine $book.Tags}}">{{$book.Title}}</A>` +
		`{{if gt (len $book.Excerpt) 0}}<DD>{{$book.Excerpt}}{{end}}{{end}}` +
		`</DL><p>`

	tpl, err := template.New("export").Funcs(funcMap).Parse(tplContent)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	// Execute template
	err = tpl.Execute(dstFile, &bookmarks)
	if err != nil {
		return errors.New(cErrorSprint(err))
	}

	fmt.Println("Export finished")
	return nil
}

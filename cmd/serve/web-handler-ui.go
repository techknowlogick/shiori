package serve

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strconv"

	"src.techknowlogick.com/shiori/utils"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
)

// serveFiles serve files
func (h *webHandler) serveFiles(c *gin.Context) {
	err := serveFile(c, c.Param("filepath"))
	utils.CheckError(err)
}

// serveIndexPage is handler for GET /
func (h *webHandler) serveIndexPage(c *gin.Context) {
	// Check token
	err := h.checkToken(c.Request)
	if err != nil {
		redirectPage(c, "/login")
		return
	}

	err = serveFile(c, "index.html")
	utils.CheckError(err)
}

// serveSubmitPage is handler for GET /submit
func (h *webHandler) serveSubmitPage(c *gin.Context) {
	err := serveFile(c, "submit.html")
	utils.CheckError(err)
}

// serveLoginPage is handler for GET /login
func (h *webHandler) serveLoginPage(c *gin.Context) {
	// Check token
	err := h.checkToken(c.Request)
	if err == nil {
		redirectPage(c, "/")
		return
	}

	err = serveFile(c, "login.html")
	utils.CheckError(err)
}

// serveBookmarkCache is handler for GET /bookmark/:id
func (h *webHandler) serveBookmarkCache(c *gin.Context) {
	// Get bookmark ID from URL
	strID := c.Param("id")
	id, err := strconv.Atoi(strID)
	utils.CheckError(err)

	// Get bookmarks in database
	bookmarks, err := h.db.GetBookmarks(true, id)
	utils.CheckError(err)

	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Create template
	funcMap := template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"hostname": func(s string) string {
			parsed, err := nurl.ParseRequestURI(s)
			if err != nil || len(parsed.Scheme) == 0 {
				return s
			}

			return parsed.Hostname()
		},
	}

	tplCache, err := createTemplate("cache.html", funcMap)
	utils.CheckError(err)

	bt, err := json.Marshal(&bookmarks[0])
	utils.CheckError(err)

	// Execute template
	strBt := string(bt)
	err = tplCache.Execute(c.Writer, &strBt)
	utils.CheckError(err)
}

// serveThumbnailImage is handler for GET /thumb/:id
func (h *webHandler) serveThumbnailImage(c *gin.Context) {
	// Get bookmark ID from URL
	id := c.Param("id")

	// Open image
	imgPath := fp.Join(h.dataDir, "thumb", id)
	img, err := os.Open(imgPath)
	utils.CheckError(err)
	defer img.Close()

	// Get image type from its 512 first bytes
	buffer := make([]byte, 512)
	_, err = img.Read(buffer)
	utils.CheckError(err)

	mimeType := http.DetectContentType(buffer)
	c.Header("Content-Type", mimeType)

	// Serve image
	img.Seek(0, 0)
	_, err = io.Copy(c.Writer, img)
	utils.CheckError(err)
}

func serveFile(c *gin.Context, path string) error {
	// Open file
	box := packr.New("views", "../../dist")
	_, fname := fp.Split(path)
	src, err := box.Find(fname)
	if err != nil {
		return err
	}

	// Get content type
	ext := fp.Ext(fname)
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		c.Header("Content-Type", mimeType)
	}

	// Serve file
	c.Writer.Write(src)
	return nil
}

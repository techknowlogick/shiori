package serve

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"src.techknowlogick.com/shiori/utils"

	valid "github.com/asaskevich/govalidator"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-shiori/go-readability"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"src.techknowlogick.com/shiori/model"
)

// login is handler for POST /api/login
func (h *webHandler) apiLogin(c *gin.Context) {
	// Decode request
	var request model.LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&request)
	utils.CheckError(err)

	// Get account data from database
	account, err := h.db.GetAccount(request.Username)
	utils.CheckError(err)

	// Compare password with database
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(request.Password))
	if err != nil {
		panic(fmt.Errorf("Username and password don't match"))
	}

	// Calculate expiration time
	nbf := time.Now()
	exp := time.Now().Add(12 * time.Hour)
	if request.Remember {
		exp = time.Now().Add(7 * 24 * time.Hour)
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": nbf.Unix(),
		"exp": exp.Unix(),
		"sub": account.ID,
	})

	tokenString, err := token.SignedString(h.jwtKey)
	utils.CheckError(err)

	// Return tokenc.Request
	fmt.Fprint(c.Writer, tokenString)
}

// apiGetBookmarks is handler for GET /api/bookmarks
func (h *webHandler) apiGetBookmarks(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Get URL queries
	keyword := c.Request.URL.Query().Get("keyword")
	strTags := c.Request.URL.Query().Get("tags")
	tags := strings.Split(strTags, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	// Fetch all matching bookmarks
	bookmarks, err := h.db.SearchBookmarks(true, keyword, tags...)
	utils.CheckError(err)

	err = json.NewEncoder(c.Writer).Encode(&bookmarks)
	utils.CheckError(err)
}

// apiGetTags is handler for GET /api/tags
func (h *webHandler) apiGetTags(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Fetch all tags
	tags, err := h.db.GetTags()
	utils.CheckError(err)

	err = json.NewEncoder(c.Writer).Encode(&tags)
	utils.CheckError(err)
}

// apiInsertBookmark is handler for POST /api/bookmark
func (h *webHandler) apiInsertBookmark(c *gin.Context) {
	// Enable CORS for this endpoint
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Decode request
	book := model.Bookmark{}
	err = json.NewDecoder(c.Request.Body).Decode(&book)
	utils.CheckError(err)

	// Make sure URL valid
	parsedURL, err := nurl.Parse(book.URL)
	if err != nil || !valid.IsRequestURL(book.URL) {
		panic(fmt.Errorf("URL is not valid"))
	}

	// Clear fragment and UTM parameters from URL
	parsedURL.Fragment = ""
	clearUTMParams(parsedURL)
	book.URL = parsedURL.String()

	// Fetch data from internet
	article, _ := readability.FromURL(parsedURL.String(), 20*time.Second)

	book.Author = article.Byline
	book.MinReadTime = int(math.Floor(float64(article.Length)/(987+188) + 0.5))
	book.MaxReadTime = int(math.Floor(float64(article.Length)/(987-188) + 0.5))
	book.Content = article.TextContent
	book.HTML = article.Content

	// If title and excerpt doesnt have submitted value, use from article
	if book.Title == "" {
		book.Title = article.Title
	}

	if book.Excerpt == "" {
		book.Excerpt = strings.Map(fixUtf, article.Excerpt)
	}

	// Make sure title is not empty
	if book.Title == "" {
		book.Title = book.URL
	}

	// Check if book has content
	if book.Content != "" {
		book.HasContent = true
	}

	// Save bookmark image to local disk
	u2, err := uuid.NewV4()
	if err != nil {
		utils.CheckError(err)
	}
	imgPath := fp.Join(h.dataDir, "thumb", u2.String())
	err = downloadFile(article.Image, imgPath, 20*time.Second)
	if err == nil {
		book.ImageURL = fmt.Sprintf("/thumb/%s", u2)
	}

	// Save bookmark to database
	err = h.db.InsertBookmark(&book)
	if err != nil {
		utils.CheckError(err)
	}

	// Return new saved result
	err = json.NewEncoder(c.Writer).Encode(&book)
	utils.CheckError(err)
}

// apiDeleteBookmarks is handler for DELETE /api/bookmark
func (h *webHandler) apiDeleteBookmark(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(c.Request.Body).Decode(&ids)
	utils.CheckError(err)

	// Delete bookmarks
	err = h.db.DeleteBookmarks(ids...)
	utils.CheckError(err)

	// Delete thumbnail image from local disk
	for _, id := range ids {
		imgPath := fp.Join(h.dataDir, "thumb", fmt.Sprintf("%d", id))
		os.Remove(imgPath)
	}

	fmt.Fprint(c.Writer, 1)
}

// apiUpdateBookmark is handler for PUT /api/bookmarks
func (h *webHandler) apiUpdateBookmark(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Decode request
	request := model.Bookmark{}
	err = json.NewDecoder(c.Request.Body).Decode(&request)
	utils.CheckError(err)

	// Validate input
	if request.Title == "" {
		panic(fmt.Errorf("Title must not empty"))
	}

	// Get existing bookmark from database
	reqID := request.ID
	bookmarks, err := h.db.GetBookmarks(true, reqID)
	utils.CheckError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Set new bookmark data
	book := bookmarks[0]
	book.Title = request.Title
	book.Excerpt = request.Excerpt

	// Set new tags
	for i := range book.Tags {
		book.Tags[i].Deleted = true
	}

	for _, newTag := range request.Tags {
		for i, oldTag := range book.Tags {
			if newTag.Name == oldTag.Name {
				newTag.ID = oldTag.ID
				book.Tags[i].Deleted = false
				break
			}
		}

		if newTag.ID == 0 {
			book.Tags = append(book.Tags, newTag)
		}
	}

	// Update database
	res, err := h.db.UpdateBookmarks(book)
	utils.CheckError(err)

	// Return new saved result
	err = json.NewEncoder(c.Writer).Encode(&res[0])
	utils.CheckError(err)
}

// apiUpdateBookmarkTags is handler for PUT /api/bookmarks/tags
func (h *webHandler) apiUpdateBookmarkTags(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Decode request
	request := struct {
		IDs  []int       `json:"ids"`
		Tags []model.Tag `json:"tags"`
	}{}

	err = json.NewDecoder(c.Request.Body).Decode(&request)
	utils.CheckError(err)

	// Validate input
	if len(request.IDs) == 0 || len(request.Tags) == 0 {
		panic(fmt.Errorf("IDs and tags must not empty"))
	}

	// Get existing bookmark from database
	bookmarks, err := h.db.GetBookmarks(true, request.IDs...)
	utils.CheckError(err)
	if len(bookmarks) == 0 {
		panic(fmt.Errorf("No bookmark with matching index"))
	}

	// Set new tags
	for i, book := range bookmarks {
		for _, newTag := range request.Tags {
			for _, oldTag := range book.Tags {
				if newTag.Name == oldTag.Name {
					newTag.ID = oldTag.ID
					break
				}
			}

			if newTag.ID == 0 {
				book.Tags = append(book.Tags, newTag)
			}
		}

		bookmarks[i] = book
	}

	// Update database
	res, err := h.db.UpdateBookmarks(bookmarks...)
	utils.CheckError(err)

	// Return new saved result
	err = json.NewEncoder(c.Writer).Encode(&res)
	utils.CheckError(err)
}

// apiUpdateCache is handler for PUT /api/cache
func (h *webHandler) apiUpdateCache(c *gin.Context) {
	// Check token
	err := h.checkAPIToken(c.Request)
	utils.CheckError(err)

	// Decode request
	ids := []int{}
	err = json.NewDecoder(c.Request.Body).Decode(&ids)
	utils.CheckError(err)

	// Prepare wait group and mutex
	wg := sync.WaitGroup{}

	// Fetch bookmarks from database
	books, err := h.db.GetBookmarks(false, ids...)
	utils.CheckError(err)

	// Download new cache data
	for _, book := range books {
		wg.Add(1)

		go func(book *model.Bookmark) {
			fmt.Println(book.ID)
			// Make sure to stop wait group
			defer wg.Done()

			// Parse URL
			parsedURL, err := nurl.Parse(book.URL)
			if err != nil || !valid.IsRequestURL(book.URL) {
				return
			}

			// Fetch data from internet
			article, err := readability.FromURL(parsedURL.String(), 20*time.Second)
			if err != nil {
				return
			}

			book.Excerpt = article.Excerpt
			book.Author = article.Byline
			book.MinReadTime = int(math.Floor(float64(article.Length)/(987+188) + 0.5))
			book.MaxReadTime = int(math.Floor(float64(article.Length)/(987-188) + 0.5))
			book.Content = article.TextContent
			book.HTML = article.Content

			// Make sure title is not empty
			if article.Title != "" {
				book.Title = article.Title
			}

			// Check if book has content
			if book.Content != "" {
				book.HasContent = true
			}

			// Update bookmark image in local disk
			u2, err := uuid.NewV4()
			if err != nil {
				utils.CheckError(err)
			}
			imgPath := fp.Join(h.dataDir, "thumb", u2.String())
			err = downloadFile(article.Image, imgPath, 20*time.Second)
			if err == nil {
				book.ImageURL = fmt.Sprintf("/thumb/%s", u2)
			}
		}(&book)
	}

	// Wait until all finished
	wg.Wait()

	// Update database
	res, err := h.db.UpdateBookmarks(books...)
	utils.CheckError(err)

	// Return new saved result
	err = json.NewEncoder(c.Writer).Encode(&res)
	utils.CheckError(err)
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

func clearUTMParams(url *nurl.URL) {
	newQuery := nurl.Values{}
	for key, value := range url.Query() {
		if !strings.HasPrefix(key, "utm_") {
			newQuery[key] = value
		}
	}

	url.RawQuery = newQuery.Encode()
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}

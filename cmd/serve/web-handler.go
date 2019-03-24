package serve

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"net/http"

	"src.techknowlogick.com/shiori/database"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr/v2"
)

// webHandler is handler for every API and routes to web page
type webHandler struct {
	db       database.Database
	dataDir  string
	jwtKey   []byte
	tplCache *template.Template
}

type handlerOptions struct {
	db        database.Database
	dataDir   string
	jwtSecret string
}

// newWebHandler returns new webHandler
func newWebHandler(options *handlerOptions) (*webHandler, error) {
	// Create JWT key
	jwtKey := make([]byte, 32)
	_, err := rand.Read(jwtKey)
	if err != nil {
		return nil, err
	}
	if len(options.jwtSecret) != 0 {
		jwtKey = []byte(options.jwtSecret)
	}

	// Create handler
	handler := &webHandler{
		db:      options.db,
		dataDir: options.dataDir,
		jwtKey:  jwtKey,
	}

	return handler, nil
}

func (h *webHandler) checkToken(r *http.Request) error {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		return fmt.Errorf("Token error: Token does not exist")
	}

	token, err := jwt.Parse(tokenCookie.Value, h.jwtKeyFunc)
	if err != nil {
		return fmt.Errorf("Token error: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	err = claims.Valid()
	if err != nil {
		return fmt.Errorf("Token error: %v", err)
	}

	return nil
}

func (h *webHandler) checkAPIToken(r *http.Request) error {
	token, err := request.ParseFromRequest(r,
		request.AuthorizationHeaderExtractor,
		h.jwtKeyFunc)
	if err != nil {
		// Try to check in cookie
		return h.checkToken(r)
	}

	claims := token.Claims.(jwt.MapClaims)
	err = claims.Valid()
	if err != nil {
		return fmt.Errorf("Token error: %v", err)
	}

	return nil
}

func (h *webHandler) jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method")
	}

	return h.jwtKey, nil
}

func createTemplate(filename string, funcMap template.FuncMap) (*template.Template, error) {
	// Open file
	box := packr.New("views", "../../dist")
	src, err := box.Find(filename)
	if err != nil {
		return nil, err
	}

	// Create template
	return template.New(filename).Delims("$|", "|$").Funcs(funcMap).Parse(string(src))
}

func redirectPage(c *gin.Context, url string) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Redirect(http.StatusMovedPermanently, url)
}

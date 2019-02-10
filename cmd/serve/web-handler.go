package serve

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gobuffalo/packr/v2"
	dt "github.com/techknowlogick/shiori/database"
)

// webHandler is handler for every API and routes to web page
type webHandler struct {
	db       dt.Database
	dataDir  string
	jwtKey   []byte
	tplCache *template.Template
}

// newWebHandler returns new webHandler
func newWebHandler(db dt.Database, dataDir string) (*webHandler, error) {
	// Create JWT key
	jwtKey := make([]byte, 32)
	_, err := rand.Read(jwtKey)
	if err != nil {
		return nil, err
	}

	// Create handler
	handler := &webHandler{
		db:      db,
		dataDir: dataDir,
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

func redirectPage(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, url, 301)
}

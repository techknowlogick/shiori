package serve

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"src.techknowlogick.com/shiori/database"
	"src.techknowlogick.com/shiori/utils"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	CmdServe = cli.Command{
		Name:  "serve",
		Usage: "Serve web app for managing bookmarks",
		Description: "Run a simple annd performant web server which serves the site for managing bookmarks." +
			"If --port flag is not used, it will use port 8080 by default.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "listen, l",
				Usage:  "Address the server listens to",
				EnvVar: "SHIORI_LISTEN_ADDRESS",
			},
			cli.StringFlag{
				Name:   "jwt-secret",
				Usage:  "JWT Secret fof session protection (Default: Randon each start)",
				EnvVar: "SHIORI_JWT_SECRET",
				Hidden: true,
			},
			cli.IntFlag{
				Name:   "port, p",
				Value:  8080,
				Usage:  "Port that used by server",
				EnvVar: "SHIORI_PORT,PORT",
			},
			cli.BoolFlag{
				Name:   "insecure-default-user",
				Usage:  "For demo service this creates a temporary default user. Very insecure, do not use this flag.",
				Hidden: true,
				EnvVar: "SHIORI_INSECURE_DEMO_USER",
			},
		},
		Action: func(c *cli.Context) error {
			db, err := getDbConnection(c)

			if err != nil {
				return errors.New(utils.CErrorSprint(err))
			}

			demoUser, _ := db.GetAccount("demo")
			if demoUser.ID == 0 && c.Bool("insecure-default-user") {
				db.CreateAccount("demo", "demo")
			}

			dataDir := c.GlobalString("data-dir")
			hdl, err := newWebHandler(&handlerOptions{db: db, dataDir: dataDir, jwtSecret: c.String("jwt-secret")})
			// Parse flags
			listenAddress := c.String("listen")
			port := c.Int("port")

			// Create router
			router := httprouter.New()

			router.GET("/dist/*filepath", hdl.serveFiles)
			router.GET("/shiori/*filepath", hdl.serveFiles)

			router.GET("/", hdl.serveIndexPage)
			router.GET("/login", hdl.serveLoginPage)
			router.GET("/bookmark/:id", hdl.serveBookmarkCache)
			router.GET("/thumb/:id", hdl.serveThumbnailImage)
			router.GET("/submit", hdl.serveSubmitPage)

			router.POST("/api/login", hdl.apiLogin)
			router.GET("/api/bookmarks", hdl.apiGetBookmarks)
			router.GET("/api/tags", hdl.apiGetTags)
			router.POST("/api/bookmarks", hdl.apiInsertBookmark)
			router.PUT("/api/cache", hdl.apiUpdateCache)
			router.PUT("/api/bookmarks", hdl.apiUpdateBookmark)
			router.PUT("/api/bookmarks/tags", hdl.apiUpdateBookmarkTags)
			router.DELETE("/api/bookmarks", hdl.apiDeleteBookmark)

			// Route for panic
			router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
				http.Error(w, fmt.Sprint(arg), 500)
			}

			// Create server
			url := fmt.Sprintf("%s:%d", listenAddress, port)
			svr := &http.Server{
				Addr:         url,
				Handler:      router,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 20 * time.Second,
			}

			// Serve app
			logrus.Infoln("Serve shiori in", url)
			return svr.ListenAndServe()
		},
	}
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func getDbConnection(c *cli.Context) (database.Database, error) {
	dbType := c.GlobalString("db-type")
	dbDsn := c.GlobalString("db-dsn")
	dataDir := c.GlobalString("data-dir")

	if dbType == "sqlite3" && dbDsn == "shiori.db" {
		dbDsn = filepath.Join(dataDir, dbDsn)
	}

	db, err := database.OpenXormDatabase(dbDsn, dbType)
	return db, err

}

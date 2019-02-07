//go:generate go run assets-generator.go

package main

import (
	"os"
	fp "path/filepath"

	"github.com/techknowlogick/shiori/cmd"
	dt "github.com/techknowlogick/shiori/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var dataDir = "."

func main() {

	// check and use mysql if env values set
	if mysqlDBName := os.Getenv("SHIORI_MYSQL_DBNAME"); mysqlDBName != "" {
		mysqlDBUser := os.Getenv("SHIORI_MYSQL_USER")
		mysqlDBPass := os.Getenv("SHIORI_MYSQL_PASS")
		mysqlDBHost := os.Getenv("SHIORI_MYSQL_HOST")
		mysqlDB, err := dt.OpenMySQLDatabase(mysqlDBHost, mysqlDBUser, mysqlDBPass, mysqlDBName)
		checkError(err)
		shioriCmd := cmd.NewShioriCmd(mysqlDB, dataDir)
		if err := shioriCmd.Execute(); err != nil {
			logrus.Fatalln(err)
		}
		return
	}

	// check and use postgresql if env values set
	if postgresqlDBName := os.Getenv("SHIORI_POSTGRESQL_DBNAME"); postgresqlDBName != "" {
		postgresqlDBUser := os.Getenv("SHIORI_POSTGRESQL_USER")
		postgresqlDBPass := os.Getenv("SHIORI_POSTGRESQL_PASS")
		postgresqlDBHost := os.Getenv("SHIORI_POSTGRESQL_HOST")
		postgresqlDB, err := dt.OpenPostgreSQLDatabase(postgresqlDBHost, postgresqlDBUser, postgresqlDBPass, postgresqlDBName)
		checkError(err)
		shioriCmd := cmd.NewShioriCmd(postgresqlDB, dataDir)
		if err := shioriCmd.Execute(); err != nil {
			logrus.Fatalln(err)
		}
		return
	}

	// Open database
	dbPath := fp.Join(dataDir, "shiori.db")
	sqliteDB, err := dt.OpenSQLiteDatabase(dbPath)
	checkError(err)

	// Start cmd
	shioriCmd := cmd.NewShioriCmd(sqliteDB, dataDir)
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

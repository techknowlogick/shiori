package main

import (
	"fmt"
	"os"
	fp "path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/techknowlogick/shiori/cmd"
	dt "github.com/techknowlogick/shiori/database"
)

var dataDir = "."

func main() {

	dbType := "sqlite3"
	dsn := fp.Join(dataDir, "shiori.db")

	// check and use mysql if env values set
	if mysqlDBName := os.Getenv("SHIORI_MYSQL_DBNAME"); mysqlDBName != "" {
		mysqlDBUser := os.Getenv("SHIORI_MYSQL_USER")
		mysqlDBPass := os.Getenv("SHIORI_MYSQL_PASS")
		mysqlDBHost := os.Getenv("SHIORI_MYSQL_HOST")
		dbType = "mysql"
		dsn = fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", mysqlDBUser, mysqlDBPass, mysqlDBHost, mysqlDBName)
	}

	// check and use postgresql if env values set
	if postgresqlDBName := os.Getenv("SHIORI_POSTGRESQL_DBNAME"); postgresqlDBName != "" {
		postgresqlDBUser := os.Getenv("SHIORI_POSTGRESQL_USER")
		postgresqlDBPass := os.Getenv("SHIORI_POSTGRESQL_PASS")
		postgresqlDBHost := os.Getenv("SHIORI_POSTGRESQL_HOST")
		dbType = "postgres"
		dsn = fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", postgresqlDBUser, postgresqlDBPass, postgresqlDBHost, postgresqlDBName)
	}

	gormDB, err := dt.OpenGORMDatabase(dsn, dbType)
	checkError(err)

	// Start cmd
	shioriCmd := cmd.NewShioriCmd(gormDB, dataDir)
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

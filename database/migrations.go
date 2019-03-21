package database

import (
	"src.techknowlogick.com/shiori/database/migration"
	"src.techknowlogick.com/xormigrate"
)

var (
	migrations = []*xormigrate.Migration{
		migration.M1,
	}
)

## Using other Databases

To use another Database there are two options. You can use parameters while calling the shiori command itself or you can set environment variables.

### Using command parameters

| Parameter | Default | Values | Description |
|----------|--------|-------|---|
| --db-type  | sqlite3 | sqlite3, mssql, mysql, postgres | since this Shiori fork uses xorm as database layer, you can use several Databases. Check out the xorm docs for more informations |
| --db-dsn | shiori.db | user:password@(localhost)/dbname?charset=utf8&parseTime=True&loc=Local | Your Database Connection string |
| --show-sql-log | false | - | Bool, if set, Shiori will output all SQL Queries to the CLI |

### Using environment Variables

| Variable | Default | Values | Description |
|----------|--------|-------|---|
| SHIORI_DBTYPE | sqlite3  | sqlite3, mssql, mysql, postgres | since this Shiori fork uses xorm as database layer, you can use several Databases. Check out the xorm docs for more informations |
| SHIORI_DSN | shiori.db | user:password@(localhost)/dbname?charset=utf8&parseTime=True&loc=Local | Your Database Connection string |
| SHIORI_SHOW_SQL | false | - | Bool, if set, Shiori will output all SQL Queries to the CLI |


## Using other Databases

To use another Database there are two options. You can use parameters while calling the shiori command itself or you can set environment variables.

### Using command parameters

| Parameter | Values | Description |
|-----------|--------|---|
| --db-type   | mssql, odbc, mysql, mymysql, postgres, pgx, sqlite3, oci8, goracle | since this Shiori fork uses xorm as database layer, you can use several Databases. Check out the xorm docs for more informations |
| --db-dsn | user:password@(localhost)/dbname?charset=utf8&parseTime=True&loc=Local | Your Database Connection string |
| --show-sql-log | - | Bool, if set, Shiori will output all SQL Queries to the CLI |

### Using environment Variables

| Variable | Values | Description |
|-----------|--------|---|
| SHIORI_DBTYPE   | mssql, odbc, mysql, mymysql, postgres, pgx, sqlite3, oci8, goracle | since this Shiori fork uses xorm as database layer, you can use several Databases. Check out the xorm docs for more informations |
| SHIORI_DSN | user:password@(localhost)/dbname?charset=utf8&parseTime=True&loc=Local | Your Database Connection string |
| SHIORI_SHOW_SQL | - | Bool, if set, Shiori will output all SQL Queries to the CLI |


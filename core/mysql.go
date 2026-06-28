package core

import _ "github.com/go-sql-driver/mysql"

var mysqlDefaults = []struct{ k, v string }{
	{"charset", "utf8mb4"},
	{"collation", "utf8mb4_0900_ai_ci"},
	{"parseTime", "true"},
	{"loc", "Local"},
}

func NewMySQL(c Config) (*Connector, error) {
	if c.WritePath != "" && c.ReadPath != "" {
		return newDual("mysql", dsn(c.WritePath, mysqlDefaults), dsn(c.ReadPath, mysqlDefaults), c)
	}
	return newSingle("mysql", dsn(c.Path, mysqlDefaults), c)
}

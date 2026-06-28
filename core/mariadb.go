package core

import _ "github.com/go-sql-driver/mysql"

var mariadbDefaults = []struct{ k, v string }{
	{"charset", "utf8mb4"},
	{"collation", "utf8mb4_general_ci"},
	{"parseTime", "true"},
	{"loc", "Local"},
}

func NewMariaDB(c Config) (*Connector, error) {
	if c.WritePath != "" && c.ReadPath != "" {
		return newDual("mysql", dsn(c.WritePath, mariadbDefaults), dsn(c.ReadPath, mariadbDefaults), c)
	}
	return newSingle("mysql", dsn(c.Path, mariadbDefaults), c)
}

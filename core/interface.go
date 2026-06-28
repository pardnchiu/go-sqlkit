package core

import "database/sql"

type target int64

const (
	SQLite target = iota
	PostgreSQL
	MySQL
	MariaDB
)

type Config struct {
	ID           string `json:"id"`
	Target       target `json:"target"`
	Path         string `json:"path"`
	ReadPath     string `json:"read_path,omitempty"`
	WritePath    string `json:"write_path,omitempty"`
	Lifetime     int    `json:"lifetime,omitempty"`
	MaxOpenConns int    `json:"max_open_conns,omitempty"`
	MaxIdleConns int    `json:"max_idle_conns,omitempty"`
}

type Connector struct {
	Read  *sql.DB
	Write *sql.DB
}

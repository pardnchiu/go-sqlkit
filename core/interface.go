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
	Target       target `json:"target"`
	Path         string `json:"path"`
	Lifetime     int    `json:"lifetime,omitempty"`
	MaxOpenConns int    `json:"max_read_conns,omitempty"`
	MaxIdleConns int    `json:"max_idle_conns,omitempty"`
}

type Connector struct {
	Read  *sql.DB
	Write *sql.DB
}

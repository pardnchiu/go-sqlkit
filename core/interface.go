package core

import "database/sql"

type Config struct {
	Path         string `json:"path"`
	Lifetime     int    `json:"lifetime,omitempty"`
	MaxOpenConns int    `json:"max_read_conns,omitempty"`
	MaxIdleConns int    `json:"max_idle_conns,omitempty"`
}

type Connector struct {
	Read  *sql.DB
	Write *sql.DB
}

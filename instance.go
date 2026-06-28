package goSqlite

import (
	"fmt"

	"github.com/pardnchiu/go-sqlite/core"
)

func New(c core.Config) (*core.Connector, error) {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 48
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 24
	}

	switch c.Target {
	case core.SQLite:
		return core.NewSQLite(c)
	default:
		return nil, fmt.Errorf("unsupported target: %d", c.Target)
	}
}

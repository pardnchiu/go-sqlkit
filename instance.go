package goSqlite

import (
	"fmt"
	"maps"
	"sync"

	"github.com/pardnchiu/go-sqlite/core"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]*core.Connector)
)

func New(c core.Config) (*core.Connector, error) {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 48
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 24
	}

	var conn *core.Connector
	var err error

	switch c.Target {
	case core.SQLite:
		conn, err = core.NewSQLite(c)
	case core.MySQL:
		conn, err = core.NewMySQL(c)
	case core.MariaDB:
		conn, err = core.NewMariaDB(c)
	case core.PostgreSQL:
		return nil, fmt.Errorf("PostgreSQL is not yet supported")
	default:
		return nil, fmt.Errorf("unsupported target: %d", c.Target)
	}

	if err != nil {
		return nil, err
	}

	if c.ID != "" {
		mu.Lock()
		registry[c.ID] = conn
		mu.Unlock()
	}

	return conn, nil
}

func ID(name string) *core.Connector {
	mu.RLock()
	defer mu.RUnlock()
	return registry[name]
}

func Targets() map[string]*core.Connector {
	mu.RLock()
	defer mu.RUnlock()

	dic := make(map[string]*core.Connector, len(registry))
	maps.Copy(dic, registry)
	return dic
}

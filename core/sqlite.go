package core

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite(c Config) (*Connector, error) {
	write, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?"+
			"cache=shared"+
			"&mode=rwc"+
			"&_journal_mode=WAL"+
			"&_busy_timeout=50000"+
			"&_synchronous=NORMAL"+
			"&_txlock=immediate", c.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to open write db: %w", err)
	}

	write.SetMaxOpenConns(1)
	write.SetMaxIdleConns(1)
	write.SetConnMaxLifetime(0)

	for _, e := range []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = -131072",
		"PRAGMA mmap_size = 268435456",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA wal_autocheckpoint = 0",
		"PRAGMA journal_size_limit = 268435456",
		"PRAGMA busy_timeout = 50000",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := write.Exec(e); err != nil {
			write.Close()
			return nil, fmt.Errorf("failed to setup pragma: %w", err)
		}
	}

	if err := write.Ping(); err != nil {
		write.Close()
		return nil, fmt.Errorf("failed to ping write db: %w", err)
	}

	read, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?"+
			"cache=shared"+
			"&mode=ro"+
			"&_query_only=1"+
			"&_journal_mode=WAL"+
			"&_busy_timeout=50000", c.Path))
	if err != nil {
		write.Close()
		return nil, fmt.Errorf("failed to open read db: %w", err)
	}

	read.SetMaxOpenConns(c.MaxOpenConns)
	read.SetMaxIdleConns(c.MaxIdleConns)

	if c.Lifetime > 0 {
		read.SetConnMaxLifetime(time.Duration(c.Lifetime) * time.Second)
	} else {
		read.SetConnMaxLifetime(2 * time.Minute)
	}

	if err := read.Ping(); err != nil {
		write.Close()
		read.Close()
		return nil, fmt.Errorf("failed to ping read db: %w", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := write.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
				fmt.Printf("checkpoint failed: %v\n", err)
			}
		}
	}()

	return &Connector{
		Read:  read,
		Write: write,
	}, nil
}

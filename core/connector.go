package core

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

func (c *Connector) Query(query string, args ...any) (*sql.Rows, error) {
	return c.Read.Query(query, args...)
}

func (c *Connector) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.Read.QueryContext(ctx, query, args...)
}

func (c *Connector) Exec(query string, args ...any) (sql.Result, error) {
	return c.Write.Exec(query, args...)
}

func (c *Connector) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.Write.ExecContext(ctx, query, args...)
}

func (c *Connector) Close() {
	if c.stop != nil {
		c.stopOnce.Do(func() { close(c.stop) })
	}

	if c.Write != nil {
		if err := c.Write.Close(); err != nil {
			slog.Error("failed to close write db",
				slog.Any("error", err))
		}
	}

	if c.Read != nil && c.Read != c.Write {
		if err := c.Read.Close(); err != nil {
			slog.Error("failed to close read db",
				slog.Any("error", err))
		}
	}
}

func dsn(raw string, defaults []struct{ k, v string }) string {
	sep := "?"
	if strings.Contains(raw, "?") {
		sep = "&"
	}

	var sb strings.Builder
	sb.WriteString(raw)

	for _, d := range defaults {
		if !strings.Contains(raw, d.k+"=") {
			sb.WriteString(sep)
			sb.WriteString(d.k)
			sb.WriteString("=")
			sb.WriteString(d.v)
			sep = "&"
		}
	}

	return sb.String()
}

func newSingle(driver, dsn string, c Config) (*Connector, error) {
	db, err := db(driver, dsn, c.MaxOpenConns, c.MaxIdleConns, lifetime(c))
	if err != nil {
		return nil, err
	}

	return &Connector{
		Read:  db,
		Write: db,
	}, nil
}

func newDual(driver, writeDSN, readDSN string, c Config) (*Connector, error) {
	write, err := db(driver, writeDSN, 10, 5, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	read, err := db(driver, readDSN, c.MaxOpenConns, c.MaxIdleConns, lifetime(c))
	if err != nil {
		write.Close()
		return nil, fmt.Errorf("read: %w", err)
	}

	return &Connector{
		Read:  read,
		Write: write,
	}, nil
}

func db(driver, dsn string, maxOpen, maxIdle int, lifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

func lifetime(c Config) time.Duration {
	if c.Lifetime > 0 {
		return time.Duration(c.Lifetime) * time.Second
	}
	return 2 * time.Minute
}

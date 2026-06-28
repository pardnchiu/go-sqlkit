package core

import (
	"context"
	"database/sql"
	"log/slog"
)

func (d *Connector) Query(key, query string, args ...any) (*sql.Rows, error) {
	return d.Read.Query(query, args...)
}

func (d *Connector) QueryContext(ctx context.Context, key, query string, args ...any) (*sql.Rows, error) {
	return d.Read.QueryContext(ctx, query, args...)
}

func (d *Connector) Exec(key, query string, args ...any) (sql.Result, error) {
	return d.Write.Exec(query, args...)
}

func (d *Connector) ExecContext(ctx context.Context, key, query string, args ...any) (sql.Result, error) {
	return d.Write.ExecContext(ctx, query, args...)
}

func (d *Connector) Close() {
	if d.Read != nil {
		if err := d.Read.Close(); err != nil {
			slog.Error("failed to close read db",
				slog.Any("error", err))
		}
	}

	if d.Write != nil {
		if err := d.Write.Close(); err != nil {
			slog.Error("failed to close write db",
				slog.Any("error", err))
		}
	}
}

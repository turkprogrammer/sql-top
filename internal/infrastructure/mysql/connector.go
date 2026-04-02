package mysql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

type Connector struct {
	db *sql.DB
}

func NewConnector() *Connector {
	return &Connector{}
}

func (c *Connector) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql connection: %w", err)
	}

	// Используем константы из domain для единообразия настроек пула
	db.SetMaxOpenConns(domain.DefaultMaxConns)
	db.SetMaxIdleConns(domain.DefaultMinConns)
	db.SetConnMaxLifetime(domain.DefaultConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	c.db = db
	return nil
}

func (c *Connector) Close(ctx context.Context) error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// DB возвращает connection для внутреннего использования.
// Не является частью публичного API — используется только adapter.
func (c *Connector) DB() *sql.DB {
	return c.db
}

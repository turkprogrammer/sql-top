package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Connector управляет подключением к PostgreSQL через connection pool.
type Connector struct {
	pool *pgxpool.Pool
}

// NewConnector создаёт новый PostgreSQL connector.
func NewConnector() *Connector {
	return &Connector{}
}

// Connect устанавливает подключение к PostgreSQL с настройками из domain.Default*.
func (c *Connector) Connect(ctx context.Context, dsn string) error {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse dsn: %w", err)
	}

	// Используем константы из domain для единообразия настроек пула
	config.MaxConns = domain.DefaultMaxConns
	config.MinConns = domain.DefaultMinConns

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	c.pool = pool
	return nil
}

func (c *Connector) Close(ctx context.Context) error {
	if c.pool != nil {
		c.pool.Close()
	}
	return nil
}

// Pool возвращает connection pool для внутреннего использования.
// Не является частью публичного API — используется только adapter.
func (c *Connector) Pool() *pgxpool.Pool {
	return c.pool
}

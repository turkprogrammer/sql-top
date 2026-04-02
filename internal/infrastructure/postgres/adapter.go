package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkprogrammer/sql-top/internal/infrastructure/base"
)

// Adapter — реализация DBProvider для PostgreSQL.
// Использует composition с base.Adapter для общей логики.
type Adapter struct {
	*base.BaseAdapter
	connector *Connector
	pool      *pgxpool.Pool
}

// NewAdapter создаёт новый PostgreSQL адаптер.
// logger передаётся через DI для соблюдения принципов Clean Architecture.
func NewAdapter(dsn string, logger *slog.Logger) (*Adapter, error) {
	connector := NewConnector()

	ctx := context.Background()
	if err := connector.Connect(ctx, dsn); err != nil {
		return nil, err
	}

	// Создаём logger для компонентов
	componentLogger := logger.With("component", "postgres.Adapter")

	return &Adapter{
		BaseAdapter: base.NewBaseAdapter(
			connector,
			NewFetcher(connector, componentLogger),
			NewExplainer(connector),
			logger,
		),
		connector: connector,
		pool:      connector.Pool(),
	}, nil
}

// KillQuery завершает запрос по PID в PostgreSQL.
func (a *Adapter) KillQuery(ctx context.Context, pid int32) error {
	query := fmt.Sprintf("SELECT pg_terminate_backend(%d)", pid)
	a.Logger().Debug("killing query", "pid", pid)
	_, err := a.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("kill query failed: %w", err)
	}
	return nil
}

// Ping проверяет подключение к PostgreSQL.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.pool.Ping(ctx)
}

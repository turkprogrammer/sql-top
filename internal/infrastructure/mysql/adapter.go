package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/turkprogrammer/sql-top/internal/infrastructure/base"
)

// Adapter — реализация DBProvider для MySQL.
// Использует composition с base.Adapter для общей логики.
type Adapter struct {
	*base.BaseAdapter
	connector *Connector
	db        *sql.DB
}

// NewAdapter создаёт новый MySQL адаптер.
// logger передаётся через DI для соблюдения принципов Clean Architecture.
func NewAdapter(dsn string, logger *slog.Logger) (*Adapter, error) {
	connector := NewConnector()

	ctx := context.Background()
	if err := connector.Connect(ctx, dsn); err != nil {
		return nil, err
	}

	return &Adapter{
		BaseAdapter: base.NewBaseAdapter(
			connector,
			NewFetcher(connector, logger.With("component", "mysql.Adapter")),
			NewExplainer(connector),
			logger,
		),
		connector: connector,
		db:        connector.DB(),
	}, nil
}

// KillQuery завершает запрос по PID в MySQL.
func (a *Adapter) KillQuery(ctx context.Context, pid int32) error {
	query := fmt.Sprintf("KILL %d", pid)
	a.Logger().Debug("killing query", "pid", pid)
	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("kill query failed: %w", err)
	}
	return nil
}

// Ping проверяет подключение к MySQL.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.db.PingContext(ctx)
}

package clickhouse

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/turkprogrammer/sql-top/internal/infrastructure/base"
)

// Adapter — реализация DBProvider для ClickHouse.
// Использует composition с base.Adapter для общей логики.
type Adapter struct {
	*base.BaseAdapter
	connector *Connector
	conn      driver.Conn
}

// NewAdapter создаёт новый ClickHouse адаптер.
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
			NewFetcher(connector, logger.With("component", "clickhouse.Adapter")),
			NewExplainer(connector),
			logger,
		),
		connector: connector,
		conn:      connector.Conn(),
	}, nil
}

// KillQuery завершает запрос по query_id в ClickHouse.
func (a *Adapter) KillQuery(ctx context.Context, pid int32) error {
	query := fmt.Sprintf("KILL QUERY WHERE query_id = '%d'", pid)
	a.Logger().Debug("killing query", "pid", pid)
	_, err := a.conn.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("kill query failed: %w", err)
	}
	return nil
}

// Ping проверяет подключение к ClickHouse.
func (a *Adapter) Ping(ctx context.Context) error {
	if err := a.conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

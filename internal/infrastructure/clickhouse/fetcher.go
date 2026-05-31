package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Fetcher реализует интерфейс QueryFetcher для ClickHouse.
// Получает активные запросы из system.processes.
type Fetcher struct {
	connector *Connector
	logger    *slog.Logger
}

// NewFetcher создаёт новый fetcher для ClickHouse.
func NewFetcher(connector *Connector, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		connector: connector,
		logger:    logger.With("component", "clickhouse.Fetcher"),
	}
}

// GetActiveQueries возвращает снапшот активных запросов из system.processes.
func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "system.processes")

	start := time.Now()
	rows, err := f.queryRows(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshot := &domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   make([]domain.Query, 0, 32),
	}

	if err := f.scanRows(rows, snapshot); err != nil {
		return nil, err
	}

	f.logger.Debug("fetched active queries", "count", len(snapshot.Queries), "duration_ms", time.Since(start).Milliseconds())
	return snapshot, nil
}

const processesSQL = `
	SELECT
		tid AS pid,
		initial_user AS usename,
		database AS datname,
		toString(query_kind) AS state,
		toString(settings) AS settings,
		query,
		elapsed,
		peak_memory_usage
	FROM system.processes
	WHERE query != ''
	ORDER BY elapsed DESC
`

func (f *Fetcher) queryRows(ctx context.Context) (driver.Rows, error) {
	rows, err := f.connector.conn.Query(ctx, processesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.processes: %w", err)
	}
	return rows, nil
}

func (f *Fetcher) scanRows(rows driver.Rows, snapshot *domain.QuerySnapshot) error {
	for rows.Next() {
		q, err := f.scanRow(rows)
		if err != nil {
			f.logger.Warn("failed to scan row", "error", err)
			continue
		}
		snapshot.Queries = append(snapshot.Queries, q)
	}
	if err := rows.Err(); err != nil {
		f.logger.Error("rows iteration error", "error", err)
		return fmt.Errorf("rows error: %w", err)
	}
	return nil
}

func (f *Fetcher) scanRow(rows driver.Rows) (domain.Query, error) {
	var q domain.Query
	var elapsed float64
	var settings, peakMemory string

	err := rows.Scan(
		&q.PID,
		&q.Username,
		&q.Datname,
		&q.State,
		&settings,
		&q.Query,
		&elapsed,
		&peakMemory,
	)
	if err != nil {
		return domain.Query{}, fmt.Errorf("failed to scan row: %w", err)
	}

	q.Duration = time.Duration(elapsed * float64(time.Second))
	if q.State == "" {
		q.State = "Query"
	}
	return q, nil
}

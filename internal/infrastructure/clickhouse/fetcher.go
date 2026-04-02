package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

type Fetcher struct {
	connector *Connector
	logger    *slog.Logger
}

func NewFetcher(connector *Connector, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		connector: connector,
		logger:    logger.With("component", "clickhouse.Fetcher"),
	}
}

func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "system.processes")

	query := `
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

	start := time.Now()
	rows, err := f.connector.conn.Query(ctx, query)
	if err != nil {
		f.logger.Error("failed to query system.processes", "error", err)
		return nil, fmt.Errorf("failed to query system.processes: %w", err)
	}
	defer rows.Close()

	snapshot := &domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   make([]domain.Query, 0),
	}

	for rows.Next() {
		var q domain.Query
		var elapsed float64
		var settings, peakMemory string

		err := rows.Scan(
			&q.PID,
			&q.Usename,
			&q.Datname,
			&q.State,
			&settings,
			&q.Query,
			&elapsed,
			&peakMemory,
		)
		if err != nil {
			// Логируем ошибку сканирования, но продолжаем обработку остальных строк
			f.logger.Warn("failed to scan row", "error", err)
			continue
		}

		q.Duration = time.Duration(elapsed * float64(time.Second))

		if q.State == "" {
			q.State = "Query"
		}

		snapshot.Queries = append(snapshot.Queries, q)
	}

	if err := rows.Err(); err != nil {
		f.logger.Error("rows iteration error", "error", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	f.logger.Debug("fetched active queries", "count", len(snapshot.Queries), "duration_ms", time.Since(start).Milliseconds())

	return snapshot, nil
}

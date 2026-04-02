package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Fetcher реализует интерфейс QueryFetcher для PostgreSQL.
// Получает активные запросы из pg_stat_activity.
type Fetcher struct {
	connector *Connector
	logger    *slog.Logger
}

// NewFetcher создаёт новый fetcher для PostgreSQL.
func NewFetcher(connector *Connector, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		connector: connector,
		logger:    logger.With("component", "postgres.Fetcher"),
	}
}

func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "pg_stat_activity")

	query := `
		SELECT
			p.pid,
			COALESCE(p.usename, '') as usename,
			COALESCE(p.datname, '') as datname,
			COALESCE(p.state, '') as state,
			COALESCE(p.wait_event_type, '') as wait_event_type,
			COALESCE(p.wait_event, '') as wait_event,
			COALESCE(p.query, '<idle>') as query,
			COALESCE(EXTRACT(EPOCH FROM (now() - p.query_start))::numeric, 0) as duration_seconds,
			p.query_start
		FROM pg_stat_activity p
		WHERE
			p.pid != pg_backend_pid()
			AND p.state IS NOT NULL
			AND p.query IS NOT NULL
		ORDER BY duration_seconds DESC
	`

	start := time.Now()
	rows, err := f.connector.pool.Query(ctx, query)
	if err != nil {
		f.logger.Error("failed to query pg_stat_activity", "error", err)
		return nil, fmt.Errorf("failed to query pg_stat_activity: %w", err)
	}
	defer rows.Close()

	snapshot := &domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   make([]domain.Query, 0),
	}

	for rows.Next() {
		var q domain.Query
		var durationSeconds float64

		err := rows.Scan(
			&q.PID,
			&q.Usename,
			&q.Datname,
			&q.State,
			&q.WaitEventType,
			&q.WaitEvent,
			&q.Query,
			&durationSeconds,
			&q.QueryStart,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		q.Duration = time.Duration(durationSeconds * float64(time.Second))
		snapshot.Queries = append(snapshot.Queries, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	f.logger.Debug("fetched active queries", "count", len(snapshot.Queries), "duration_ms", time.Since(start).Milliseconds())

	return snapshot, nil
}

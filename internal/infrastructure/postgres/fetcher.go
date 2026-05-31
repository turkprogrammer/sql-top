package postgres

import (
	"context"
	"errors"
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

// GetActiveQueries возвращает снапшот активных запросов из pg_stat_activity.
func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "pg_stat_activity")

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

const activeQueriesSQL = `
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

func (f *Fetcher) queryRows(ctx context.Context) (pgx.Rows, error) {
	rows, err := f.connector.pool.Query(ctx, activeQueriesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query pg_stat_activity: %w", err)
	}
	return rows, nil
}

func (f *Fetcher) scanRows(rows pgx.Rows, snapshot *domain.QuerySnapshot) error {
	for rows.Next() {
		q, err := f.scanRow(rows)
		if err != nil {
			return err
		}
		snapshot.Queries = append(snapshot.Queries, q)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}
	return nil
}

func (f *Fetcher) scanRow(rows pgx.Rows) (domain.Query, error) {
	var q domain.Query
	var durationSeconds float64

	err := rows.Scan(
		&q.PID,
		&q.Username,
		&q.Datname,
		&q.State,
		&q.WaitEventType,
		&q.WaitEvent,
		&q.Query,
		&durationSeconds,
		&q.QueryStart,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Query{}, nil
		}
		return domain.Query{}, fmt.Errorf("failed to scan row: %w", err)
	}

	q.Duration = time.Duration(durationSeconds * float64(time.Second))
	return q, nil
}

package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Fetcher реализует интерфейс QueryFetcher для MySQL.
// Получает активные запросы из information_schema.PROCESSLIST.
type Fetcher struct {
	connector *Connector
	logger    *slog.Logger
}

// NewFetcher создаёт новый fetcher для MySQL.
func NewFetcher(connector *Connector, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		connector: connector,
		logger:    logger.With("component", "mysql.Fetcher"),
	}
}

// GetActiveQueries возвращает снапшот активных запросов из information_schema.PROCESSLIST.
func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "information_schema.PROCESSLIST")

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

const processlistSQL = `
	SELECT
		ID,
		USER,
		DB,
		COMMAND,
		STATE,
		INFO,
		TIME,
		TIME_MS
	FROM information_schema.PROCESSLIST
	WHERE COMMAND != 'Sleep'
		AND ID != CONNECTION_ID()
	ORDER BY TIME DESC
`

func (f *Fetcher) queryRows(ctx context.Context) (*sql.Rows, error) {
	rows, err := f.connector.db.QueryContext(ctx, processlistSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query processlist: %w", err)
	}
	return rows, nil
}

func (f *Fetcher) scanRows(rows *sql.Rows, snapshot *domain.QuerySnapshot) error {
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

func (f *Fetcher) scanRow(rows *sql.Rows) (domain.Query, error) {
	var q domain.Query
	var command, state sql.NullString
	var info sql.NullString
	var timeMs sql.NullInt64

	err := rows.Scan(
		&q.PID,
		&q.Username,
		&q.Datname,
		&command,
		&state,
		&info,
		&q.Duration,
		&timeMs,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Query{}, nil
		}
		return domain.Query{}, fmt.Errorf("failed to scan row: %w", err)
	}

	if state.Valid {
		q.State = state.String
	}
	if info.Valid {
		q.Query = info.String
	}

	q.State = f.resolveState(command, q.State)
	f.applyDurationFromMs(&q, timeMs)

	return q, nil
}

func (f *Fetcher) resolveState(command sql.NullString, currentState string) string {
	if !command.Valid {
		return currentState
	}
	switch command.String {
	case "Query":
		if currentState == "" {
			return "executing"
		}
	case "Binlog Dump":
		return "replication"
	case "Connect":
		return "connecting"
	}
	return currentState
}

func (f *Fetcher) applyDurationFromMs(q *domain.Query, timeMs sql.NullInt64) {
	if timeMs.Valid && q.Duration == 0 {
		q.Duration = time.Duration(timeMs.Int64) * time.Millisecond
	}
}

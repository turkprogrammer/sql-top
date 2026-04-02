package mysql

import (
	"context"
	"database/sql"
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
		logger:    logger.With("component", "mysql.Fetcher"),
	}
}

func (f *Fetcher) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	f.logger.Debug("fetching active queries", "query", "information_schema.PROCESSLIST")

	query := `
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

	start := time.Now()
	rows, err := f.connector.db.QueryContext(ctx, query)
	if err != nil {
		f.logger.Error("failed to query processlist", "error", err)
		return nil, fmt.Errorf("failed to query processlist: %w", err)
	}
	defer rows.Close()

	snapshot := &domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   make([]domain.Query, 0),
	}

	for rows.Next() {
		var q domain.Query
		var command, state sql.NullString
		var info sql.NullString
		var timeMs sql.NullInt64

		err := rows.Scan(
			&q.PID,
			&q.Usename,
			&q.Datname,
			&command,
			&state,
			&info,
			&q.Duration,
			&timeMs,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if state.Valid {
			q.State = state.String
		}

		if info.Valid {
			q.Query = info.String
		}

		if command.Valid {
			switch command.String {
			case "Query":
				if q.State == "" {
					q.State = "executing"
				}
			case "Binlog Dump":
				q.State = "replication"
			case "Connect":
				q.State = "connecting"
			}
		}

		if timeMs.Valid && q.Duration == 0 {
			q.Duration = time.Duration(timeMs.Int64) * time.Millisecond
		}

		snapshot.Queries = append(snapshot.Queries, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	f.logger.Debug("fetched active queries", "count", len(snapshot.Queries), "duration_ms", time.Since(start).Milliseconds())

	return snapshot, nil
}

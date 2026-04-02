package domain

import (
	"context"
	"time"
)

// Query представляет активный запрос в СУБД.
type Query struct {
	PID           int32
	Usename       string
	Datname       string
	State         string
	WaitEventType string
	WaitEvent     string
	Query         string
	Duration      time.Duration
	QueryStart    time.Time
	IsNew         bool
}

// QuerySnapshot представляет снапшот активных запросов на момент времени.
type QuerySnapshot struct {
	Timestamp time.Time
	Queries   []Query
}

// ExplainResult представляет результат выполнения EXPLAIN.
type ExplainResult struct {
	Plan string
}

// DBProvider — интерфейс для взаимодействия с базой данных.
// Реализуется адаптерами для PostgreSQL, MySQL и ClickHouse.
type DBProvider interface {
	// Connect устанавливает подключение к БД.
	Connect(ctx context.Context, dsn string) error
	// Close закрывает подключение к БД.
	Close(ctx context.Context) error
	// GetActiveQueries возвращает снапшот активных запросов.
	GetActiveQueries(ctx context.Context) (*QuerySnapshot, error)
	// KillQuery завершает выполнение запроса по PID.
	KillQuery(ctx context.Context, pid int32) error
	// ExplainQuery возвращает план выполнения запроса.
	ExplainQuery(ctx context.Context, query string) (*ExplainResult, error)
	// Ping проверяет подключение к БД.
	Ping(ctx context.Context) error
}

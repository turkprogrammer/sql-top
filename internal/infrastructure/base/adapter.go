package base

import (
	"context"
	"log/slog"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

// DBConnector — интерфейс для подключения к БД
type DBConnector interface {
	Connect(ctx context.Context, dsn string) error
	Close(ctx context.Context) error
}

// QueryFetcher — интерфейс для получения активных запросов
type QueryFetcher interface {
	GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error)
}

// QueryExplainer — интерфейс для EXPLAIN запросов
type QueryExplainer interface {
	ExplainQuery(ctx context.Context, query string) (string, error)
}

// QueryKiller — интерфейс для Kill запросов
type QueryKiller interface {
	KillQuery(ctx context.Context, pid int32) error
}

// Pinger — интерфейс для проверки подключения
type Pinger interface {
	Ping(ctx context.Context) error
}

// BaseAdapter — базовая реализация адаптера с общей логикой.
// Использует composition для делегирования специфичных методов.
type BaseAdapter struct {
	connector DBConnector
	fetcher   QueryFetcher
	explainer QueryExplainer
	logger    *slog.Logger
}

// Logger возвращает logger для использования в наследниках.
func (a *BaseAdapter) Logger() *slog.Logger {
	return a.logger
}

// NewBaseAdapter создаёт базовый адаптер с зависимостями.
// logger передаётся через DI для соблюдения принципов Clean Architecture.
func NewBaseAdapter(
	connector DBConnector,
	fetcher QueryFetcher,
	explainer QueryExplainer,
	logger *slog.Logger,
) *BaseAdapter {
	return &BaseAdapter{
		connector: connector,
		fetcher:   fetcher,
		explainer: explainer,
		logger:    logger.With("component", "base.Adapter"),
	}
}

// Connect подключается к БД через connector.
func (a *BaseAdapter) Connect(ctx context.Context, dsn string) error {
	a.logger.Debug("connecting to database", "dsn", sanitizeDSN(dsn))
	return a.connector.Connect(ctx, dsn)
}

// Close закрывает подключение к БД.
func (a *BaseAdapter) Close(ctx context.Context) error {
	a.logger.Debug("closing database connection")
	return a.connector.Close(ctx)
}

// GetActiveQueries возвращает снапшот активных запросов.
func (a *BaseAdapter) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	a.logger.Debug("fetching active queries")
	return a.fetcher.GetActiveQueries(ctx)
}

// ExplainQuery возвращает план выполнения запроса.
func (a *BaseAdapter) ExplainQuery(ctx context.Context, query string) (*domain.ExplainResult, error) {
	a.logger.Debug("explaining query", "query_length", len(query))
	plan, err := a.explainer.ExplainQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	return &domain.ExplainResult{Plan: plan}, nil
}

// sanitizeDSN скрывает пароль из DSN для безопасного логирования.
func sanitizeDSN(dsn string) string {
	// Простая эвристика: скрываем всё после :@
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == ':' && i+1 < len(dsn) && dsn[i+1] != '/' {
			j := i + 1
			for j < len(dsn) && dsn[j] != '@' {
				j++
			}
			if j < len(dsn) {
				return dsn[:i+1] + "****" + dsn[j:]
			}
		}
	}
	return dsn
}

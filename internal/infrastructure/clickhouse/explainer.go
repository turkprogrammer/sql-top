package clickhouse

import (
	"context"
	"fmt"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Explainer реализует интерфейс QueryExplainer для ClickHouse.
type Explainer struct {
	connector *Connector
}

// NewExplainer создаёт новый explainer для ClickHouse.
func NewExplainer(connector *Connector) *Explainer {
	return &Explainer{connector: connector}
}

// ExplainQuery возвращает план выполнения запроса для ClickHouse.
func (e *Explainer) ExplainQuery(ctx context.Context, query string) (string, error) {
	// Обрезаем длинный запрос для безопасного отображения в EXPLAIN
	safeQuery := domain.SanitizeQuery(query, domain.MaxQueryPreviewLength, false)

	explainQuery := fmt.Sprintf("EXPLAIN QUERY TREE %s", safeQuery)

	var plan string
	err := e.connector.conn.QueryRow(ctx, explainQuery).Scan(&plan)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	return plan, nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Explainer реализует интерфейс QueryExplainer для PostgreSQL.
// Предоставляет безопасный EXPLAIN без выполнения запроса.
type Explainer struct {
	connector *Connector
}

// NewExplainer создаёт новый explainer для PostgreSQL.
func NewExplainer(connector *Connector) *Explainer {
	return &Explainer{connector: connector}
}

// ExplainQuery возвращает план выполнения запроса в формате JSON.
// Использует SanitizeQuery для безопасной обработки длинных запросов.
func (e *Explainer) ExplainQuery(ctx context.Context, query string) (string, error) {
	// Обрезаем длинный запрос для безопасного отображения в EXPLAIN
	safeQuery := domain.SanitizeQuery(query, domain.MaxQueryPreviewLength, true)

	explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", safeQuery)

	var planJSON string
	err := e.connector.pool.QueryRow(ctx, explainQuery).Scan(&planJSON)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	return planJSON, nil
}

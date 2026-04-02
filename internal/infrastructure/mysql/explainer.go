package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

type Explainer struct {
	connector *Connector
}

func NewExplainer(connector *Connector) *Explainer {
	return &Explainer{connector: connector}
}

func (e *Explainer) ExplainQuery(ctx context.Context, query string) (string, error) {
	// Обрезаем длинный запрос для безопасного отображения в EXPLAIN
	safeQuery := domain.SanitizeQuery(query, domain.MaxQueryPreviewLength, false)

	explainQuery := fmt.Sprintf("EXPLAIN %s", safeQuery)

	var id, selectType, table string
	var partition sql.NullString
	var type_, possibleKeys, key string
	var keyLen, ref sql.NullString
	var rows, filtered int
	var extra string

	err := e.connector.db.QueryRowContext(ctx, explainQuery).Scan(
		&id, &selectType, &table, &partition,
		&type_, &possibleKeys, &key, &keyLen, &ref,
		&rows, &filtered, &extra,
	)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	plan := fmt.Sprintf(`{
  "id": %s,
  "select_type": "%s",
  "table": "%s",
  "type": "%s",
  "possible_keys": "%s",
  "key": "%s",
  "key_len": "%s",
  "ref": "%s",
  "rows": %d,
  "filtered": %d,
  "Extra": "%s"
}`,
		id, selectType, table, type_, possibleKeys, key, keyLen.String, ref.String, rows, filtered, extra)

	return plan, nil
}

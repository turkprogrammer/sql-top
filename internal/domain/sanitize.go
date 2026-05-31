package domain

import "strings"

// MaxQueryPreviewLength — максимальная длина запроса для предпросмотра
const MaxQueryPreviewLength = 500

// SanitizeQuery обрезает запрос до maxLength символов.
// Если addEllipsis=true и запрос длиннее maxLength, добавляется "..." в конце.
// Используется для безопасного отображения длинных SQL-запросов в EXPLAIN.
func SanitizeQuery(query string, maxLength int, addEllipsis bool) string {
	if len(query) <= maxLength {
		return query
	}

	if addEllipsis {
		return query[:maxLength] + "..."
	}

	return query[:maxLength]
}

// SanitizeDSN скрывает пароль из DSN для безопасного логирования.
func SanitizeDSN(dsn string) string {
	if idx := strings.Index(dsn, "://"); idx != -1 {
		rest := dsn[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			return dsn[:idx+3] + "***@" + rest[atIdx+1:]
		}
	}
	return dsn
}

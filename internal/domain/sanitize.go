package domain

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

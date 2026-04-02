package domain

import (
	"strings"
	"testing"
)

// TestSanitizeQuery_Short проверяет обработку коротких запросов (без обрезки).
func TestSanitizeQuery_Short(t *testing.T) {
	query := "SELECT * FROM users"
	result := SanitizeQuery(query, MaxQueryPreviewLength, true)

	if result != query {
		t.Errorf("ожидалось %q, получено %q", query, result)
	}
}

// TestSanitizeQuery_Long проверяет обрезку длинных запросов с добавлением троеточия.
func TestSanitizeQuery_Long(t *testing.T) {
	query := strings.Repeat("SELECT * FROM users WHERE id = ", 20)
	result := SanitizeQuery(query, MaxQueryPreviewLength, true)

	if len(result) > MaxQueryPreviewLength+3 {
		t.Errorf("длина должна быть ≤ %d, получено %d", MaxQueryPreviewLength+3, len(result))
	}

	if !strings.HasSuffix(result, "...") {
		t.Error("длинный запрос должен заканчиваться на '...'")
	}
}

// TestSanitizeQuery_Empty проверяет обработку пустой строки.
func TestSanitizeQuery_Empty(t *testing.T) {
	query := ""
	result := SanitizeQuery(query, MaxQueryPreviewLength, true)

	if result != "" {
		t.Errorf("ожидалась пустая строка, получено %q", result)
	}
}

// TestSanitizeQuery_WithoutEllipsis проверяет обрезку без добавления троеточия.
func TestSanitizeQuery_WithoutEllipsis(t *testing.T) {
	query := strings.Repeat("SELECT * FROM users WHERE id = ", 20)
	result := SanitizeQuery(query, MaxQueryPreviewLength, false)

	if len(result) > MaxQueryPreviewLength {
		t.Errorf("длина должна быть ≤ %d, получено %d", MaxQueryPreviewLength, len(result))
	}

	if strings.HasSuffix(result, "...") {
		t.Error("запрос без ellipsis не должен заканчиваться на '...'")
	}
}

// TestSanitizeQuery_ExactlyMaxLength проверяет запрос точно длиной в maxLength.
func TestSanitizeQuery_ExactlyMaxLength(t *testing.T) {
	query := strings.Repeat("a", MaxQueryPreviewLength)
	result := SanitizeQuery(query, MaxQueryPreviewLength, true)

	if len(result) != MaxQueryPreviewLength {
		t.Errorf("ожидалась длина %d, получено %d", MaxQueryPreviewLength, len(result))
	}

	if result != query {
		t.Error("ровный запрос не должен обрезаться")
	}
}

// TestSanitizeQuery_OneCharOver проверяет запрос на один символ длиннее maxLength.
func TestSanitizeQuery_OneCharOver(t *testing.T) {
	query := strings.Repeat("a", MaxQueryPreviewLength+1)
	result := SanitizeQuery(query, MaxQueryPreviewLength, true)

	expectedLen := MaxQueryPreviewLength + 3 // + "..."
	if len(result) != expectedLen {
		t.Errorf("ожидалась длина %d, получено %d", expectedLen, len(result))
	}
}

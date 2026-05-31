package ui

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// TestHandleKey_EnterShowsExplain проверяет открытие EXPLAIN modal по Enter.
func TestHandleKey_EnterShowsExplain(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{PID: 100, Query: "SELECT * FROM users"},
	}
	model.selectedIndex = 0

	// Эмулируем нажатие Enter
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	// Ждём завершения горутины (showExplainModal выполняется асинхронно)
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что открылся explain modal
	m.mu.Lock()
	showExplain := m.showExplain
	m.mu.Unlock()

	if !showExplain {
		t.Error("ожидалось showExplain = true")
	}
}

// TestHandleKey_YCopiesQuery проверяет копирование запроса по клавише 'y'.
func TestHandleKey_YCopiesQuery(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{PID: 100, Query: "SELECT * FROM users"},
	}
	model.selectedIndex = 0

	// Эмулируем нажатие 'y'
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	// Ждём завершения горутины (copyQueryToClipboard выполняется асинхронно)
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что появился copy confirm
	m.mu.Lock()
	copyConfirm := m.copyConfirm
	m.mu.Unlock()

	if !copyConfirm {
		t.Error("ожидалось copyConfirm = true")
	}
}

// TestHandleKey_KShowsKillConfirm проверяет запрос подтверждения Kill Query по 'k'.
func TestHandleKey_KShowsKillConfirm(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{PID: 100, Query: "SELECT * FROM users"},
	}
	model.selectedIndex = 0

	// Эмулируем нажатие 'k'
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	// Проверяем, что появился kill confirm
	if !m.killConfirm {
		t.Error("ожидалось killConfirm = true")
	}
	if m.killPID != 100 {
		t.Errorf("ожидался killPID = 100, получен %d", m.killPID)
	}
}

// TestHandleNavigation_Up проверяет навигацию вверх по списку запросов.
func TestHandleNavigation_Up(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{PID: 100},
		{PID: 200},
		{PID: 300},
	}
	model.selectedIndex = 2

	// Эмулируем нажатие Up
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	if m.selectedIndex != 1 {
		t.Errorf("ожидался selectedIndex = 1, получен %d", m.selectedIndex)
	}
}

// TestHandleNavigation_Down проверяет навигацию вниз по списку запросов.
func TestHandleNavigation_Down(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{PID: 100},
		{PID: 200},
		{PID: 300},
	}
	model.selectedIndex = 0

	// Эмулируем нажатие Down
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	if m.selectedIndex != 1 {
		t.Errorf("ожидался selectedIndex = 1, получен %d", m.selectedIndex)
	}
}

// TestHandleKillConfirm_Confirm проверяет подтверждение Kill Query по 'y'.
func TestHandleKillConfirm_Confirm(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.killConfirm = true
	model.killPID = 123

	// Используем handleKey вместо handleKillConfirm напрямую
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	// Проверяем, что kill confirm сброшен
	if m.killConfirm {
		t.Error("ожидалось killConfirm = false")
	}
}

// TestHandleKillConfirm_Cancel проверяет отмену Kill Query по 'n'.
func TestHandleKillConfirm_Cancel(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.killConfirm = true
	model.killPID = 123

	// Используем handleKey вместо handleKillConfirm напрямую
	newModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m, ok := newModel.(*Model)
	if !ok {
		t.Fatal("не удалось привести модель к *Model")
	}

	// Проверяем, что kill confirm сброшен
	if m.killConfirm {
		t.Error("ожидалось killConfirm = false")
	}
	if m.killPID != 0 {
		t.Errorf("ожидался killPID = 0, получен %d", m.killPID)
	}
}

// TestRenderTableRow проверяет рендеринг строки таблицы с данными запроса.
func TestRenderTableRow(t *testing.T) {
	provider := &mockProvider{}
	logger := slog.Default()
	model := NewModel(provider, logger)
	model.queries = []domain.Query{
		{
			PID:      12345,
			Username:  "postgres",
			Datname:  "testdb",
			State:    "active",
			Query:    "SELECT * FROM users WHERE id = 1",
			Duration: 500 * time.Millisecond,
		},
	}
	model.selectedIndex = 0

	// Рендерим строку
	row := model.renderTableRow(model.queries[0], 0)

	// Проверяем, что строка содержит PID
	if !strings.Contains(row, "12345") {
		t.Errorf("строка должна содержать PID: %s", row)
	}

	// Проверяем, что строка содержит username
	if !strings.Contains(row, "postgres") {
		t.Errorf("строка должна содержать username: %s", row)
	}
}

// TestTruncateQuery_Short проверяет обрезку короткого запроса (без изменений).
func TestTruncateQuery_Short(t *testing.T) {
	query := "SELECT * FROM users"
	result := truncateQuery(query)

	if result != query {
		t.Errorf("ожидалось %q, получено %q", query, result)
	}
}

// TestTruncateQuery_Long проверяет обрезку длинного запроса с троеточием.
func TestTruncateQuery_Long(t *testing.T) {
	query := strings.Repeat("SELECT * FROM users WHERE id = ", 10)
	result := truncateQuery(query)

	if len(result) > 63 {
		t.Errorf("длина должна быть ≤ 63, получено %d", len(result))
	}

	if !strings.HasSuffix(result, "...") {
		t.Error("длинный запрос должен заканчиваться на '...'")
	}
}

// TestFormatDuration_Milliseconds проверяет форматирование длительности в миллисекундах.
func TestFormatDuration_Milliseconds(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "менее секунды",
			duration: 500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "1 секунда",
			duration: 1 * time.Second,
			expected: "1.0s",
		},
		{
			name:     "30 секунд",
			duration: 30 * time.Second,
			expected: "30.0s",
		},
		{
			name:     "1 минута",
			duration: 60 * time.Second,
			expected: "1.0m",
		},
		{
			name:     "2.5 минуты",
			duration: 150 * time.Second,
			expected: "2.5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, ожидалось %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// TestGetRowStyle_Active проверяет стиль для активного запроса.
func TestGetRowStyle_Active(t *testing.T) {
	query := domain.Query{
		State:         "active",
		WaitEventType: "",
	}

	style := getRowStyle(query, false)

	// Проверяем, что стиль не пустой (проверка через рендер)
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("ожидался непустой рендер")
	}
}

// TestGetRowStyle_Waiting проверяет стиль для запроса в ожидании (Lock).
func TestGetRowStyle_Waiting(t *testing.T) {
	query := domain.Query{
		State:         "active",
		WaitEventType: "Lock",
	}

	style := getRowStyle(query, false)

	// Проверяем, что стиль не пустой
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("ожидался непустой рендер")
	}
}

// TestGetRowStyle_Selected проверяет стиль для выбранного запроса.
func TestGetRowStyle_Selected(t *testing.T) {
	query := domain.Query{
		State:         "idle",
		WaitEventType: "",
	}

	style := getRowStyle(query, true)

	// Для выбранной строки должен быть рендер с фоном
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("ожидался непустой рендер для выбранной строки")
	}
}

// TestTruncate_Shorter проверяет усечение строки, которая короче лимита.
func TestTruncate_Shorter(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("ожидалось 'hello', получено %q", result)
	}
}

// TestTruncate_Longer проверяет усечение длинной строки с добавлением троеточия.
func TestTruncate_Longer(t *testing.T) {
	result := truncate("hello world", 8)
	if len(result) > 8 {
		t.Errorf("длина должна быть ≤ 8, получено %d", len(result))
	}
	if !strings.HasSuffix(result, "..") {
		t.Error("должно заканчиваться на '..'")
	}
}

// TestRenderFooter проверяет рендеринг нижнего колонтитула с информацией о подключении.
func TestRenderFooter(t *testing.T) {
	footer := renderFooter()

	if footer == "" {
		t.Error("ожидался непустой footer")
	}

	if !strings.Contains(footer, "Navigate") {
		t.Error("footer должен содержать 'Navigate'")
	}

	if !strings.Contains(footer, "EXPLAIN") {
		t.Error("footer должен содержать 'EXPLAIN'")
	}
}

// mockProvider — заглушка для domain.DBProvider
type mockProvider struct{}

func (m *mockProvider) Connect(ctx context.Context, dsn string) error {
	return nil
}

func (m *mockProvider) GetActiveQueries(ctx context.Context) (*domain.QuerySnapshot, error) {
	return &domain.QuerySnapshot{Queries: []domain.Query{}}, nil
}

func (m *mockProvider) ExplainQuery(ctx context.Context, query string) (*domain.ExplainResult, error) {
	return &domain.ExplainResult{Plan: "Seq Scan on users"}, nil
}

func (m *mockProvider) KillQuery(ctx context.Context, pid int32) error {
	return nil
}

func (m *mockProvider) Ping(ctx context.Context) error {
	return nil
}

func (m *mockProvider) Close(ctx context.Context) error {
	return nil
}

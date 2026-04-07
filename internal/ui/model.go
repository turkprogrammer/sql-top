package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/turkprogrammer/sql-top/internal/domain"
	"github.com/turkprogrammer/sql-top/internal/history"
)

type Model struct {
	provider      domain.DBProvider
	diffEngine    *domain.DiffEngine
	history       *history.RingBuffer
	queries       []domain.Query
	selectedIndex int
	err           error
	connected     bool
	showExplain   bool
	explainPlan   string
	explainQuery  string
	killConfirm   bool
	killPID       int32
	copyConfirm   bool
	width         int
	height        int
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	logger        *slog.Logger
}

func NewModel(provider domain.DBProvider, logger *slog.Logger) *Model {
	ctx, cancel := context.WithCancel(context.Background())
	return &Model{
		provider:   provider,
		diffEngine: domain.NewDiffEngine(),
		history:    history.NewRingBuffer(domain.DefaultRingBufferCapacity),
		queries:    make([]domain.Query, 0),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger.With("component", "ui.Model"),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchLoop(),
		m.pingLoop(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case domain.QuerySnapshot:
		return m.handleSnapshot(msg)
	case error:
		m.err = msg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Обработка режима EXPLAIN
	if m.showExplain {
		return m.handleExplainMode(msg)
	}

	// Обработка подтверждения Kill
	if m.killConfirm {
		return m.handleKillConfirm(msg)
	}

	// Основная навигация и действия
	return m.handleNavigation(msg)
}

// handleExplainMode обрабатывает клавиши в режиме просмотра EXPLAIN
func (m *Model) handleExplainMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
		m.showExplain = false
		m.explainPlan = ""
	}
	return m, nil
}

// handleKillConfirm обрабатывает подтверждение Kill query
func (m *Model) handleKillConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		switch strings.ToLower(string(msg.Runes)) {
		case "y":
			pid := m.killPID
			m.killConfirm = false
			m.killPID = 0
			// Запускаем в отдельной горутине, чтобы не блокировать UI
			go m.executeKill(pid)
		case "n", "q":
			m.killConfirm = false
			m.killPID = 0
		}
	case tea.KeyEsc, tea.KeyCtrlC:
		m.killConfirm = false
		m.killPID = 0
	}
	return m, nil
}

// handleNavigation обрабатывает навигацию и основные действия
func (m *Model) handleNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.cancel()
		return m, tea.Quit

	case tea.KeyUp:
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case tea.KeyDown:
		if m.selectedIndex < len(m.queries)-1 {
			m.selectedIndex++
		}

	case tea.KeyEnter:
		if len(m.queries) > 0 && m.selectedIndex < len(m.queries) {
			query := m.queries[m.selectedIndex].Query
			// Запускаем в отдельной горутине, чтобы не блокировать UI
			go m.showExplainModal(query)
		}

	case tea.KeyRunes:
		switch strings.ToLower(string(msg.Runes)) {
		case "q":
			m.cancel()
			return m, tea.Quit

		case "k":
			if len(m.queries) > 0 && m.selectedIndex < len(m.queries) {
				m.killConfirm = true
				m.killPID = m.queries[m.selectedIndex].PID
			}

		case "y":
			if len(m.queries) > 0 && m.selectedIndex < len(m.queries) {
				query := m.queries[m.selectedIndex].Query
				// Запускаем в отдельной горутине, чтобы не блокировать UI
				go m.copyQueryToClipboard(query)
			}
		}
	}

	return m, nil
}

// copyQueryToClipboard копирует запрос в буфер обмена.
// Использует github.com/atotto/clipboard для кроссплатформенной поддержки.
func (m *Model) copyQueryToClipboard(query string) {
	err := clipboard.WriteAll(query)

	m.mu.Lock()
	if err != nil {
		m.err = fmt.Errorf("failed to copy query: %w", err)
		m.mu.Unlock()
		return
	}

	m.copyConfirm = true
	m.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(m.ctx, domain.GetClipboardConfirmTimeout())
		defer cancel()

		<-ctx.Done()

		m.mu.Lock()
		m.copyConfirm = false
		m.mu.Unlock()
	}()
}

func (m *Model) handleSnapshot(snapshot domain.QuerySnapshot) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.diffEngine.MarkNewQueries(&snapshot)
	m.history.Push(snapshot)
	m.queries = snapshot.Queries
	m.err = nil
	return m, nil
}

func (m *Model) fetchLoop() tea.Cmd {
	return func() tea.Msg {
		ticker := time.NewTicker(domain.DefaultPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				m.logger.Debug("fetch loop stopped", "reason", "context done")
				return nil
			case <-ticker.C:
				snapshot, err := m.provider.GetActiveQueries(m.ctx)
				if err != nil {
					m.logger.Error("failed to fetch active queries", "error", err)
					return fmt.Errorf("fetch queries: %w", err)
				}
				m.logger.Debug("fetched queries", "count", len(snapshot.Queries))
				return *snapshot
			}
		}
	}
}

func (m *Model) pingLoop() tea.Cmd {
	return func() tea.Msg {
		ticker := time.NewTicker(domain.GetPingInterval())
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				m.logger.Debug("ping loop stopped", "reason", "context done")
				return nil
			case <-ticker.C:
				if err := m.provider.Ping(m.ctx); err != nil {
					m.logger.Error("failed to ping database", "error", err)
					return fmt.Errorf("ping database: %w", err)
				}
				m.mu.Lock()
				m.connected = true
				m.mu.Unlock()
			}
		}
	}
}

// executeKill выполняет завершение запроса с использованием KillQueryTimeout.
func (m *Model) executeKill(pid int32) {
	ctx, cancel := context.WithTimeout(m.ctx, domain.GetKillQueryTimeout())
	defer cancel()

	if err := m.provider.KillQuery(ctx, pid); err != nil {
		m.mu.Lock()
		m.err = fmt.Errorf("failed to kill query PID %d: %w", pid, err)
		m.mu.Unlock()
	}
}

// showExplainModal показывает модальное окно с EXPLAIN с использованием ExplainQueryTimeout.
func (m *Model) showExplainModal(query string) {
	m.mu.Lock()
	m.explainQuery = query
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(m.ctx, domain.GetExplainQueryTimeout())
	defer cancel()

	result, err := m.provider.ExplainQuery(ctx, query)

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.err = fmt.Errorf("failed to explain: %w", err)
		m.showExplain = false
	} else {
		m.explainPlan = result.Plan
		m.showExplain = true
	}
}

func (m *Model) View() string {
	var sb strings.Builder

	if !m.connected {
		sb.WriteString(StyleWarning.Render("Connecting..."))
		sb.WriteString("\n")
		return sb.String()
	}

	sb.WriteString(renderHeader())
	sb.WriteString("\n")
	sb.WriteString(m.renderTable())
	sb.WriteString("\n")
	sb.WriteString(renderFooter())

	if m.showExplain {
		sb.WriteString(m.renderExplainModal())
	}

	if m.killConfirm {
		sb.WriteString(m.renderKillConfirm())
	}

	if m.copyConfirm {
		sb.WriteString("\n")
		sb.WriteString(StyleActive.Render("Query copied to clipboard"))
	}

	if m.err != nil {
		sb.WriteString("\n")
		sb.WriteString(StyleError.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return sb.String()
}

func renderHeader() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	return header.Render(" SQL-Top | Live Query Monitor ")
}

// renderTable рендерит таблицу запросов
func (m *Model) renderTable() string {
	if len(m.queries) == 0 {
		return StyleHelp.Render("No active queries")
	}

	var sb strings.Builder

	sb.WriteString(m.renderTableHeader())
	sb.WriteString(m.renderTableSeparator())
	sb.WriteString(m.renderTableRows())

	return sb.String()
}

// renderTableHeader рендерит заголовок таблицы
func (m *Model) renderTableHeader() string {
	header := fmt.Sprintf("%-6s %-10s %-10s %-8s %-10s %s\n",
		"PID", "User", "DB", "Duration", "Wait", "Query")
	return StyleHeader.Render(header)
}

// renderTableSeparator рендерит разделитель.
// Использует ширину окна или значение по умолчанию.
func (m *Model) renderTableSeparator() string {
	sepLen := domain.DefaultTableColumnWidth
	if m.width > 0 {
		sepLen = m.width
	}
	return strings.Repeat("─", sepLen) + "\n"
}

// renderTableRows рендерит строки таблицы
func (m *Model) renderTableRows() string {
	var sb strings.Builder

	for i, q := range m.queries {
		sb.WriteString(m.renderTableRow(q, i))
		if i < len(m.queries)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderTableRow рендерит одну строку таблицы
func (m *Model) renderTableRow(q domain.Query, index int) string {
	style := getRowStyle(q, index == m.selectedIndex)
	duration := formatDuration(q.Duration)
	waitEvent := q.WaitEvent
	if waitEvent == "" {
		waitEvent = "-"
	}
	queryPreview := truncateQuery(q.Query)

	return fmt.Sprintf("%-6d %-10s %-10s %-8s %-10s %s",
		q.PID,
		truncate(q.Usename, 10),
		truncate(q.Datname, 10),
		style.Render(duration),
		waitEvent,
		style.Render(queryPreview),
	)
}

// truncateQuery обрезает запрос до максимальной длины.
func truncateQuery(query string) string {
	if len(query) > domain.DefaultQueryTruncateLength {
		return query[:domain.DefaultQueryTruncateLength] + "..."
	}
	return query
}

// getRowStyle возвращает стиль для строки таблицы
func getRowStyle(q domain.Query, selected bool) lipgloss.Style {
	if selected {
		return lipgloss.NewStyle().Background(lipgloss.Color("240"))
	}

	// Проверяем состояние ожидания
	if domain.WaitEventType(q.WaitEventType).IsWaiting() {
		return StyleWaiting
	}

	// Проверяем состояние запроса
	switch q.State {
	case "active", "executing", "Query":
		return StyleActive
	case "idle", "idle in transaction":
		return StyleIdle
	default:
		return lipgloss.NewStyle()
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-2] + ".."
	}
	return s
}

func renderFooter() string {
	return helpStyle.Render("↑↓ Navigate | Enter EXPLAIN | K Kill | y Copy | q Quit")
}

func (m *Model) renderExplainModal() string {
	width := domain.DefaultModalWidth
	height := domain.DefaultModalHeight
	if m.width > 0 {
		width = m.width - 4
	}
	if m.height > 0 {
		height = m.height - 6
	}

	modalStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width).
		Height(height)

	content := StyleModalTitle.Render("EXPLAIN (JSON)\n\n") + m.explainPlan

	return "\n" + modalStyle.Render(content) + "\n" + helpStyle.Render("Press ESC to close")
}

func (m *Model) renderKillConfirm() string {
	return fmt.Sprintf("\n%s\n",
		StyleKillConfirm.Render(
			fmt.Sprintf(" Terminate backend PID %d? (y/n) ", m.killPID),
		),
	)
}

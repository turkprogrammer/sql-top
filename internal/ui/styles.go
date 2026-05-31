package ui

import "github.com/charmbracelet/lipgloss"

var (
	// StyleActive для активных запросов (active/executing).
	// Цвет: ярко-зелёный.
	StyleActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	// StyleWaiting для запросов в ожидании (lock, IO).
	// Цвет: оранжевый.
	StyleWaiting = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	// StyleIdle для неактивных соединений (idle, idle in transaction).
	// Цвет: серый.
	StyleIdle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// StyleHeader для заголовка таблицы.
	// Стиль: жирный, белый цвет.
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	// StyleError для ошибок.
	// Цвет: красный.
	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	// StyleWarning для предупреждений.
	// Цвет: оранжевый.
	StyleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	// StyleHelp для справки.
	// Цвет: серый.
	StyleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// StyleModalTitle для заголовков модальных окон.
	// Стиль: жирный, белый цвет.
	StyleModalTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	// StyleKillConfirm для подтверждения удаления запроса.
	// Стиль: красный фон, белый текст, жирный.
	StyleKillConfirm = lipgloss.NewStyle().
				Background(lipgloss.Color("52")).
				Foreground(lipgloss.Color("15")).
				Bold(true)

	// styleSelected для выделенной строки таблицы.
	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("240"))

	// styleDefault для строк с неизвестным состоянием.
	styleDefault = lipgloss.NewStyle()

	// styleHeaderBar для верхней панели заголовка.
	styleHeaderBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)

package domain

import "time"

// Конфигурационные константы для подключения к БД
const (
	// DefaultMaxConns — максимальное количество подключений в пуле
	DefaultMaxConns = 2

	// DefaultMinConns — минимальное количество подключений в пуле
	DefaultMinConns = 1

	// DefaultConnMaxLifetime — максимальное время жизни подключения
	DefaultConnMaxLifetime = 5 * time.Minute
)

// Конфигурационные константы для polling
const (
	// DefaultPollInterval — интервал опроса активных запросов
	DefaultPollInterval = 1 * time.Second

	// DefaultRingBufferCapacity — ёмкость буфера истории запросов
	DefaultRingBufferCapacity = 5000
)

// Конфигурационные константы для ClickHouse
const (
	// DefaultClickHousePort — порт по умолчанию для ClickHouse
	DefaultClickHousePort = "9000"

	// DefaultClickHouseDatabase — база данных по умолчанию для ClickHouse
	DefaultClickHouseDatabase = "default"

	// DefaultClickHouseUser — пользователь по умолчанию для ClickHouse
	DefaultClickHouseUser = "default"
)

// Константы UI
const (
	// DefaultTableColumnWidth — ширина колонки таблицы по умолчанию
	DefaultTableColumnWidth = 120

	// ClipboardFileName — временный файл для копирования запроса
	ClipboardFileName = "clipboard.tmp"

	// DefaultModalWidth — ширина модального окна по умолчанию
	DefaultModalWidth = 80

	// DefaultModalHeight — высота модального окна по умолчанию
	DefaultModalHeight = 30

	// DefaultQueryTruncateLength — максимальная длина запроса перед обрезкой
	DefaultQueryTruncateLength = 60
)

// Константы timeout
const (
	// ClipboardConfirmTimeout — время подтверждения копирования
	ClipboardConfirmTimeout = 2 * time.Second

	// PingInterval — интервал ping проверки
	PingInterval = 5 * time.Second

	// KillQueryTimeout — timeout для kill query
	KillQueryTimeout = 5 * time.Second

	// ExplainQueryTimeout — timeout для explain query
	ExplainQueryTimeout = 10 * time.Second
)

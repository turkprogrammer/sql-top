package domain

import (
	"os"
	"time"
)

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

	// DefaultModalWidth — ширина модального окна по умолчанию
	DefaultModalWidth = 80

	// DefaultModalHeight — высота модального окна по умолчанию
	DefaultModalHeight = 30

	// DefaultQueryTruncateLength — максимальная длина запроса перед обрезкой
	DefaultQueryTruncateLength = 60
)

// Константы timeout (значения по умолчанию)
const (
	// DefaultClipboardConfirmTimeout — время подтверждения копирования
	DefaultClipboardConfirmTimeout = 2 * time.Second

	// DefaultPingInterval — интервал ping проверки
	DefaultPingInterval = 5 * time.Second

	// DefaultKillQueryTimeout — timeout для kill query
	DefaultKillQueryTimeout = 5 * time.Second

	// DefaultExplainQueryTimeout — timeout для explain query
	DefaultExplainQueryTimeout = 10 * time.Second
)

// GetClipboardConfirmTimeout возвращает настраиваемый timeout для подтверждения копирования.
// Можно переопределить через env: SQLTOP_CLIPBOARD_TIMEOUT
func GetClipboardConfirmTimeout() time.Duration {
	return getEnvDuration("SQLTOP_CLIPBOARD_TIMEOUT", DefaultClipboardConfirmTimeout)
}

// GetPingInterval возвращает настраиваемый интервал ping проверки.
// Можно переопределить через env: SQLTOP_PING_INTERVAL
func GetPingInterval() time.Duration {
	return getEnvDuration("SQLTOP_PING_INTERVAL", DefaultPingInterval)
}

// GetKillQueryTimeout возвращает настраиваемый timeout для kill query.
// Можно переопределить через env: SQLTOP_KILL_TIMEOUT
func GetKillQueryTimeout() time.Duration {
	return getEnvDuration("SQLTOP_KILL_TIMEOUT", DefaultKillQueryTimeout)
}

// GetExplainQueryTimeout возвращает настраиваемый timeout для explain query.
// Можно переопределить через env: SQLTOP_EXPLAIN_TIMEOUT
func GetExplainQueryTimeout() time.Duration {
	return getEnvDuration("SQLTOP_EXPLAIN_TIMEOUT", DefaultExplainQueryTimeout)
}

// Для обратной совместимости - используем геттеры
var (
	ClipboardConfirmTimeout = GetClipboardConfirmTimeout()
	PingInterval            = GetPingInterval()
	KillQueryTimeout        = GetKillQueryTimeout()
	ExplainQueryTimeout     = GetExplainQueryTimeout()
)

// getEnvDuration читает duration из переменной окружения или возвращает значение по умолчанию.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultValue
}

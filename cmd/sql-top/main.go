package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbletea"
	"github.com/turkprogrammer/sql-top/internal/domain"
	"github.com/turkprogrammer/sql-top/internal/infrastructure/clickhouse"
	"github.com/turkprogrammer/sql-top/internal/infrastructure/mysql"
	"github.com/turkprogrammer/sql-top/internal/infrastructure/postgres"
	"github.com/turkprogrammer/sql-top/internal/ui"
)

// Version устанавливается при сборке через -ldflags "-X main.Version=x.y.z"
var Version = "dev"

// Config содержит конфигурацию приложения
type Config struct {
	DSN    string
	Logger *slog.Logger
}

// parseArgs парсит аргументы командной строки
func parseArgs(args []string) (*Config, error) {
	flagSet, dsn := createFlagSet()

	if err := parseFlagSet(flagSet, args, dsn); err != nil {
		return nil, err
	}

	logger := createLogger()
	return &Config{DSN: *dsn, Logger: logger}, nil
}

// createFlagSet создаёт и настраивает флагSet
func createFlagSet() (*flag.FlagSet, *string) {
	flagSet := flag.NewFlagSet("sql-top", flag.ExitOnError)
	dsn := flagSet.String("dsn", "", "Database connection string (PostgreSQL, MySQL, or ClickHouse)")

	flagSet.Usage = func() {
		fmt.Printf("SQL-Top v%s: Live Query Monitor for PostgreSQL, MySQL & ClickHouse\n", Version)
		fmt.Println()
		fmt.Println("Usage: sql-top [options] [dsn]")
		fmt.Println()
		fmt.Println("Options:")
		flagSet.PrintDefaults()
		fmt.Println("  -version")
		fmt.Println("    	Show version information")
		fmt.Println()
		fmt.Println("Supported databases:")
		fmt.Println("  PostgreSQL:  postgres://user:pass@host:5432/db")
		fmt.Println("  MySQL:      mysql://user:pass@host:3306/db")
		fmt.Println("  ClickHouse: clickhouse://user:pass@host:9000/db")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  sql-top postgres://user:pass@localhost:5432/mydb")
		fmt.Println("  sql-top mysql://user:pass@localhost:3306/mydb")
		fmt.Println("  sql-top clickhouse://user:pass@localhost:9000/mydb")
		fmt.Println("  sql-top -dsn postgres://user:pass@localhost:5432/mydb")
		fmt.Println("  sql-top -version")
	}

	return flagSet, dsn
}

// parseFlagSet обрабатывает аргументы и возвращает DSN
func parseFlagSet(flagSet *flag.FlagSet, args []string, dsn *string) error {
	if len(args) == 0 {
		flagSet.Usage()
		os.Exit(1)
	}

	for i, arg := range args {
		if arg == "--help" || arg == "-h" {
			flagSet.Usage()
			os.Exit(0)
		}

		if arg == "--version" || arg == "-version" {
			fmt.Printf("sql-top version %s\n", Version)
			os.Exit(0)
		}

		if arg == "--dsn" || arg == "-d" {
			if i+1 < len(args) {
				*dsn = args[i+1]
				break
			}
		}

		if !flagSet.Parsed() && !strings.HasPrefix(arg, "-") {
			*dsn = arg
		}
	}

	flagSet.Parse(args)
	if flagSet.NFlag() == 0 && flagSet.NArg() > 0 {
		*dsn = flagSet.Arg(0)
	}

	if *dsn == "" {
		return fmt.Errorf("DSN is required")
	}

	return nil
}

// createLogger создаёт структурированный логгер
func createLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("SQLTOP_DEBUG") == "1" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Убираем время из логов для чистоты вывода в TUI
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
}

// createAdapter создает адаптер для нужной БД
func createAdapter(ctx context.Context, dsn string, logger *slog.Logger) (domain.DBProvider, error) {
	dbType := detectDBType(dsn)

	logger.Debug("определение типа БД", "type", dbType, "dsn", sanitizeDSN(dsn))

	switch dbType {
	case "postgres":
		logger.Info("подключение к PostgreSQL")
		return postgres.NewAdapter(dsn, logger)
	case "mysql":
		logger.Info("подключение к MySQL")
		return mysql.NewAdapter(dsn, logger)
	case "clickhouse":
		logger.Info("подключение к ClickHouse")
		return clickhouse.NewAdapter(dsn, logger)
	default:
		return nil, fmt.Errorf("unknown database type; supported: postgres, mysql, clickhouse")
	}
}

// sanitizeDSN скрывает пароль из DSN для логирования
func sanitizeDSN(dsn string) string {
	// Простая санитизация — убираем пароль
	if idx := strings.Index(dsn, "://"); idx != -1 {
		rest := dsn[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx != -1 {
			return dsn[:idx+3] + "***@" + rest[atIdx+1:]
		}
	}
	return dsn
}

// detectDBType определяет тип БД по DSN
func detectDBType(dsn string) string {
	dsnLower := strings.ToLower(dsn)
	if strings.HasPrefix(dsnLower, "postgres://") || strings.HasPrefix(dsnLower, "postgresql://") {
		return "postgres"
	}
	if strings.HasPrefix(dsnLower, "mysql://") {
		return "mysql"
	}
	if strings.HasPrefix(dsnLower, "clickhouse://") {
		return "clickhouse"
	}
	return "unknown"
}

func main() {
	config, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Usage: sql-top [dsn]")
		os.Exit(1)
	}

	// Создаём контекст с поддержкой graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обрабатываем сигналы завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	adapter, err := createAdapter(ctx, config.DSN, config.Logger)
	if err != nil {
		config.Logger.Error("ошибка подключения", "error", err)
		os.Exit(1)
	}
	defer adapter.Close(ctx)

	if err := adapter.Ping(ctx); err != nil {
		config.Logger.Error("ошибка ping", "error", err)
		os.Exit(1)
	}

	config.Logger.Info("подключение успешно", "connected", true)

	m := ui.NewModel(adapter, config.Logger)

	// Запускаем программу с обработкой сигналов
	go func() {
		<-sigChan
		config.Logger.Info("получен сигнал завершения")
		cancel()
	}()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		config.Logger.Error("ошибка выполнения программы", "error", err)
		os.Exit(1)
	}

	config.Logger.Info("программа завершена")
}

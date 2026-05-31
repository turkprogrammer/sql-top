package clickhouse

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// Connector управляет подключением к ClickHouse.
type Connector struct {
	conn driver.Conn
}

// NewConnector создаёт новый ClickHouse connector.
func NewConnector() *Connector {
	return &Connector{}
}

// Connect устанавливает подключение к ClickHouse с настройками из DSN.
func (c *Connector) Connect(ctx context.Context, dsn string) error {
	opts, err := parseDSN(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse clickhouse dsn: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	c.conn = conn
	return nil
}

// Close закрывает подключение к ClickHouse.
func (c *Connector) Close(ctx context.Context) error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Conn возвращает connection для внутреннего использования.
// Не является частью публичного API — используется только adapter.
func (c *Connector) Conn() driver.Conn {
	return c.conn
}

func parseDSN(dsn string) (*clickhouse.Options, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}
	port := u.Port()
	if port == "" {
		port = domain.DefaultClickHousePort
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		database = domain.DefaultClickHouseDatabase
	}

	username := u.User.Username()
	if username == "" {
		username = domain.DefaultClickHouseUser
	}
	password, hasPassword := u.User.Password()
	if !hasPassword {
		password = ""
	}

	return &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	}, nil
}

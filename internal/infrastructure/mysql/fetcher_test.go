package mysql

import (
	"log/slog"
	"testing"
)

// TestConnector_NewConnector проверяет создание нового коннектора.
func TestConnector_NewConnector(t *testing.T) {
	connector := NewConnector()

	if connector == nil {
		t.Fatal("NewConnector вернул nil")
	}

	if connector.db != nil {
		t.Error("ожидалось nil db для нового коннектора")
	}
}

// TestNewExplainer проверяет создание explainer.
func TestNewExplainer(t *testing.T) {
	connector := &Connector{}
	explainer := NewExplainer(connector)

	if explainer == nil {
		t.Fatal("NewExplainer вернул nil")
	}

	if explainer.connector != connector {
		t.Error("explainer.connector не установлен")
	}
}

// TestNewFetcher проверяет создание fetcher.
func TestNewFetcher(t *testing.T) {
	connector := &Connector{}
	logger := slog.Default()
	fetcher := NewFetcher(connector, logger)

	if fetcher == nil {
		t.Fatal("NewFetcher вернул nil")
	}

	if fetcher.connector != connector {
		t.Error("fetcher.connector не установлен")
	}
}

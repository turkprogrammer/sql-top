package domain

import (
	"testing"
	"time"
)

// TestDiffEngine_NewQueries проверяет обнаружение новых запросов.
func TestDiffEngine_NewQueries(t *testing.T) {
	tests := []struct {
		name            string
		previousPIDs    map[int32]struct{}
		snapshot        *QuerySnapshot
		expectedNewPIDs map[int32]bool
	}{
		{
			name:         "пустой снапшот",
			previousPIDs: map[int32]struct{}{},
			snapshot: &QuerySnapshot{
				Timestamp: time.Now(),
				Queries:   []Query{},
			},
			expectedNewPIDs: map[int32]bool{},
		},
		{
			name:         "первый снапшот с запросами",
			previousPIDs: map[int32]struct{}{},
			snapshot: &QuerySnapshot{
				Timestamp: time.Now(),
				Queries: []Query{
					{PID: 100, Query: "SELECT 1"},
					{PID: 200, Query: "SELECT 2"},
				},
			},
			expectedNewPIDs: map[int32]bool{
				100: true,
				200: true,
			},
		},
		{
			name: "смешанные запросы (новые и существующие)",
			previousPIDs: map[int32]struct{}{
				100: {},
			},
			snapshot: &QuerySnapshot{
				Timestamp: time.Now(),
				Queries: []Query{
					{PID: 100, Query: "SELECT 1"}, // существующий
					{PID: 200, Query: "SELECT 2"}, // новый
					{PID: 300, Query: "SELECT 3"}, // новый
				},
			},
			expectedNewPIDs: map[int32]bool{
				100: false,
				200: true,
				300: true,
			},
		},
		{
			name: "запросы завершены (PID удалены)",
			previousPIDs: map[int32]struct{}{
				100: {},
				200: {},
			},
			snapshot: &QuerySnapshot{
				Timestamp: time.Now(),
				Queries: []Query{
					{PID: 100, Query: "SELECT 1"},
				},
			},
			expectedNewPIDs: map[int32]bool{
				100: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &DiffEngine{
				previousPIDs: tt.previousPIDs,
			}

			engine.MarkNewQueries(tt.snapshot)

			for _, q := range tt.snapshot.Queries {
				expectedNew := tt.expectedNewPIDs[q.PID]
				if q.IsNew != expectedNew {
					t.Errorf("PID %d: IsNew = %v, expected %v", q.PID, q.IsNew, expectedNew)
				}
			}
		})
	}
}

// TestDiffEngine_StatePersistence проверяет сохранение состояния между вызовами.
func TestDiffEngine_StatePersistence(t *testing.T) {
	engine := NewDiffEngine()

	// Первый снапшот
	snapshot1 := &QuerySnapshot{
		Timestamp: time.Now(),
		Queries: []Query{
			{PID: 100, Query: "SELECT 1"},
			{PID: 200, Query: "SELECT 2"},
		},
	}
	engine.MarkNewQueries(snapshot1)

	// Проверяем, что все запросы помечены как новые
	for _, q := range snapshot1.Queries {
		if !q.IsNew {
			t.Errorf("PID %d: ожидалось IsNew = true", q.PID)
		}
	}

	// Второй снапшот (те же PID)
	snapshot2 := &QuerySnapshot{
		Timestamp: time.Now().Add(1 * time.Second),
		Queries: []Query{
			{PID: 100, Query: "SELECT 1"},
			{PID: 200, Query: "SELECT 2"},
		},
	}
	engine.MarkNewQueries(snapshot2)

	// Проверяем, что запросы НЕ помечены как новые
	for _, q := range snapshot2.Queries {
		if q.IsNew {
			t.Errorf("PID %d: ожидалось IsNew = false (существующий запрос)", q.PID)
		}
	}
}

// TestDiffEngine_CompleteReplacement проверяет полную замену запросов.
func TestDiffEngine_CompleteReplacement(t *testing.T) {
	engine := NewDiffEngine()

	// Первый снапшот
	snapshot1 := &QuerySnapshot{
		Timestamp: time.Now(),
		Queries: []Query{
			{PID: 100, Query: "SELECT 1"},
		},
	}
	engine.MarkNewQueries(snapshot1)

	// Второй снапшот (полная замена)
	snapshot2 := &QuerySnapshot{
		Timestamp: time.Now().Add(1 * time.Second),
		Queries: []Query{
			{PID: 200, Query: "SELECT 2"},
			{PID: 300, Query: "SELECT 3"},
		},
	}
	engine.MarkNewQueries(snapshot2)

	// Все запросы во втором снапшоте должны быть новыми
	if len(snapshot2.Queries) != 2 {
		t.Fatalf("ожидалось 2 запроса, получено %d", len(snapshot2.Queries))
	}

	for _, q := range snapshot2.Queries {
		if !q.IsNew {
			t.Errorf("PID %d: ожидалось IsNew = true (полная замена)", q.PID)
		}
	}
}

// TestNewDiffEngine проверяет создание нового движка diff.
func TestNewDiffEngine(t *testing.T) {
	engine := NewDiffEngine()

	if engine == nil {
		t.Fatal("NewDiffEngine вернул nil")
	}

	if engine.previousPIDs == nil {
		t.Error("previousPIDs не инициализирован")
	}

	if len(engine.previousPIDs) != 0 {
		t.Errorf("ожидалась пустая карта, длина = %d", len(engine.previousPIDs))
	}
}

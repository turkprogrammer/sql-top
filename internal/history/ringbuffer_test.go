package history

import (
	"testing"
	"time"

	"github.com/turkprogrammer/sql-top/internal/domain"
)

// TestRingBuffer_PushAndGet проверяет добавление и получение снапшотов.
func TestRingBuffer_PushAndGet(t *testing.T) {
	buffer := NewRingBuffer(5)

	snapshots := []domain.QuerySnapshot{
		{Timestamp: time.Now(), Queries: []domain.Query{{PID: 1}}},
		{Timestamp: time.Now().Add(1 * time.Second), Queries: []domain.Query{{PID: 2}}},
		{Timestamp: time.Now().Add(2 * time.Second), Queries: []domain.Query{{PID: 3}}},
	}

	for _, s := range snapshots {
		buffer.Push(s)
	}

	result := buffer.GetAll()

	if len(result) != 3 {
		t.Errorf("ожидалось 3 снапшота, получено %d", len(result))
	}

	for i, s := range result {
		if s.Queries[0].PID != int32(i+1) {
			t.Errorf("снапшот %d: ожидался PID %d, получен %d", i, i+1, s.Queries[0].PID)
		}
	}
}

// TestRingBuffer_CapacityOverflow проверяет переполнение буфера и перезапись старых данных.
func TestRingBuffer_CapacityOverflow(t *testing.T) {
	buffer := NewRingBuffer(3)

	for i := 0; i < 5; i++ {
		buffer.Push(domain.QuerySnapshot{
			Timestamp: time.Now(),
			Queries:   []domain.Query{{PID: int32(i)}},
		})
	}

	result := buffer.GetAll()

	if len(result) != 3 {
		t.Errorf("ожидалось 3 снапшота (емкость), получено %d", len(result))
	}

	// Проверяем, что старые данные перезаписаны (остались последние 3)
	expectedPIDs := []int32{2, 3, 4}
	for i, s := range result {
		if s.Queries[0].PID != expectedPIDs[i] {
			t.Errorf("снапшот %d: ожидался PID %d, получен %d", i, expectedPIDs[i], s.Queries[0].PID)
		}
	}
}

// TestRingBuffer_Latest проверяет получение последнего снапшота.
func TestRingBuffer_Latest(t *testing.T) {
	buffer := NewRingBuffer(5)

	// Пустой буфер
	latest := buffer.Latest()
	if latest != nil {
		t.Error("ожидалось nil для пустого буфера")
	}

	// Один элемент
	buffer.Push(domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   []domain.Query{{PID: 100}},
	})

	latest = buffer.Latest()
	if latest == nil {
		t.Fatal("ожидался снапшот, получено nil")
	}
	if latest.Queries[0].PID != 100 {
		t.Errorf("ожидался PID 100, получен %d", latest.Queries[0].PID)
	}

	// Несколько элементов
	buffer.Push(domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   []domain.Query{{PID: 200}},
	})
	buffer.Push(domain.QuerySnapshot{
		Timestamp: time.Now(),
		Queries:   []domain.Query{{PID: 300}},
	})

	latest = buffer.Latest()
	if latest.Queries[0].PID != 300 {
		t.Errorf("ожидался PID 300 (последний), получен %d", latest.Queries[0].PID)
	}
}

// TestRingBuffer_Len проверяет подсчёт количества снапшотов в буфере.
func TestRingBuffer_Len(t *testing.T) {
	buffer := NewRingBuffer(5)

	if buffer.Len() != 0 {
		t.Errorf("ожидалась длина 0, получено %d", buffer.Len())
	}

	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 1}}})
	if buffer.Len() != 1 {
		t.Errorf("ожидалась длина 1, получено %d", buffer.Len())
	}

	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 2}}})
	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 3}}})
	if buffer.Len() != 3 {
		t.Errorf("ожидалась длина 3, получено %d", buffer.Len())
	}

	// Переполнение
	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 4}}})
	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 5}}})
	buffer.Push(domain.QuerySnapshot{Timestamp: time.Now(), Queries: []domain.Query{{PID: 6}}})

	if buffer.Len() != 5 {
		t.Errorf("ожидалась длина 5 (емкость), получено %d", buffer.Len())
	}
}

// TestRingBuffer_EmptyGetAll проверяет получение данных из пустого буфера.
func TestRingBuffer_EmptyGetAll(t *testing.T) {
	buffer := NewRingBuffer(5)

	result := buffer.GetAll()

	if result != nil {
		t.Errorf("ожидалось nil для пустого буфера, получено %v", result)
	}
}

// TestRingBuffer_ExactCapacity проверяет работу буфера при заполнении до точной ёмкости.
func TestRingBuffer_ExactCapacity(t *testing.T) {
	buffer := NewRingBuffer(3)

	for i := 0; i < 3; i++ {
		buffer.Push(domain.QuerySnapshot{
			Timestamp: time.Now(),
			Queries:   []domain.Query{{PID: int32(i)}},
		})
	}

	result := buffer.GetAll()
	if len(result) != 3 {
		t.Errorf("ожидалось 3 снапшота, получено %d", len(result))
	}

	expectedPIDs := []int32{0, 1, 2}
	for i, s := range result {
		if s.Queries[0].PID != expectedPIDs[i] {
			t.Errorf("снапшот %d: ожидался PID %d, получен %d", i, expectedPIDs[i], s.Queries[0].PID)
		}
	}
}

// TestRingBuffer_OrderPreservation проверяет сохранение порядка снапшотов.
func TestRingBuffer_OrderPreservation(t *testing.T) {
	buffer := NewRingBuffer(10)

	// Добавляем 7 снапшотов
	for i := 0; i < 7; i++ {
		buffer.Push(domain.QuerySnapshot{
			Timestamp: time.Unix(int64(i), 0),
			Queries:   []domain.Query{{PID: int32(i)}},
		})
	}

	result := buffer.GetAll()

	// Проверяем порядок (от старого к новому)
	for i, s := range result {
		if s.Timestamp.Unix() != int64(i) {
			t.Errorf("снапшот %d: ожидался timestamp %d, получен %d", i, i, s.Timestamp.Unix())
		}
	}
}

// TestNewRingBuffer проверяет создание нового кольцевого буфера.
func TestNewRingBuffer(t *testing.T) {
	capacity := 100
	buffer := NewRingBuffer(capacity)

	if buffer == nil {
		t.Fatal("NewRingBuffer вернул nil")
	}

	if buffer.Len() != 0 {
		t.Errorf("ожидалась начальная длина 0, получено %d", buffer.Len())
	}

	latest := buffer.Latest()
	if latest != nil {
		t.Error("ожидалось nil для нового буфера")
	}
}

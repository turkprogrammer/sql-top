package history

import (
	"github.com/turkprogrammer/sql-top/internal/domain"
)

// RingBuffer реализует кольцевой буфер для хранения истории снапшотов запросов.
// При переполнении старые записи перезаписываются новыми.
type RingBuffer struct {
	buffer   []domain.QuerySnapshot
	capacity int
	index    int
	count    int
}

// NewRingBuffer создаёт новый кольцевой буфер с указанной ёмкостью.
// Если capacity <= 0, устанавливается ёмкость по умолчанию (1).
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBuffer{
		buffer:   make([]domain.QuerySnapshot, capacity),
		capacity: capacity,
		index:    0,
		count:    0,
	}
}

// Push добавляет новый снапшот в буфер.
// Если буфер заполнен, перезаписывает самый старый снапшот.
func (rb *RingBuffer) Push(snapshot domain.QuerySnapshot) {
	rb.buffer[rb.index] = snapshot
	rb.index = (rb.index + 1) % rb.capacity
	if rb.count < rb.capacity {
		rb.count++
	}
}

// GetAll возвращает все снапшоты в порядке добавления (от старых к новым).
func (rb *RingBuffer) GetAll() []domain.QuerySnapshot {
	if rb.count == 0 {
		return nil
	}

	result := make([]domain.QuerySnapshot, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.index - rb.count + i + rb.capacity) % rb.capacity
		result[i] = rb.buffer[idx]
	}
	return result
}

// Latest возвращает последний добавленный снапшот.
// Возвращает nil, если буфер пуст.
// Возвращает defensive copy для защиты внутреннего состояния.
func (rb *RingBuffer) Latest() *domain.QuerySnapshot {
	if rb.count == 0 {
		return nil
	}
	idx := (rb.index - 1 + rb.capacity) % rb.capacity
	snap := rb.buffer[idx]
	return &snap
}

// Len возвращает количество снапшотов в буфере.
func (rb *RingBuffer) Len() int {
	return rb.count
}

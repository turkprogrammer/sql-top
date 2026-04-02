package domain

// QueryState представляет состояние запроса в СУБД.
type QueryState string

const (
	// StateActive — запрос выполняется.
	StateActive QueryState = "active"
	// StateIdle — соединение простаивает.
	StateIdle QueryState = "idle"
	// StateIdleInTransaction — соединение простаивает в транзакции.
	StateIdleInTransaction QueryState = "idle in transaction"
	// StateFastpathFunctionCall — выполняется fastpath функция.
	StateFastpathFunctionCall QueryState = "fastpath function call"
	// StateDisabled — состояние отключено.
	StateDisabled QueryState = "disabled"
)

// WaitEventType представляет тип события ожидания.
type WaitEventType string

const (
	// WaitEventLock — ожидание блокировки.
	WaitEventLock WaitEventType = "Lock"
	// WaitEventBufferPin — ожидание буфера.
	WaitEventBufferPin WaitEventType = "BufferPin"
	// WaitEventIO — ожидание I/O операции.
	WaitEventIO WaitEventType = "IO"
	// WaitEventLWLock — ожидание лёгкой блокировки.
	WaitEventLWLock WaitEventType = "LWLock"
	// WaitEventExtension — ожидание расширения.
	WaitEventExtension WaitEventType = "Extension"
	// WaitEventIPC — ожидание IPC.
	WaitEventIPC WaitEventType = "IPC"
	// WaitEventTimeout — ожидание таймаута.
	WaitEventTimeout WaitEventType = "Timeout"
	// WaitEventNone — нет ожидания.
	WaitEventNone WaitEventType = ""
)

// IsWaiting возвращает true, если запрос ожидает событие.
func (w WaitEventType) IsWaiting() bool {
	return w != "" && w != WaitEventNone
}

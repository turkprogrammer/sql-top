package domain

// DiffEngine отслеживает новые запросы между снапшотами.
// Используется для подсветки вновь появившихся запросов в UI.
type DiffEngine struct {
	previousPIDs map[int32]struct{}
}

// NewDiffEngine создаёт новый движок для отслеживания изменений.
func NewDiffEngine() *DiffEngine {
	return &DiffEngine{
		previousPIDs: make(map[int32]struct{}),
	}
}

// MarkNewQueries помечает новые запросы в снапшоте флагом IsNew.
// Сравнивает PID запросов с предыдущим снапшотом.
func (d *DiffEngine) MarkNewQueries(snapshot *QuerySnapshot) {
	currentPIDs := make(map[int32]struct{})

	for i := range snapshot.Queries {
		q := &snapshot.Queries[i]
		currentPIDs[q.PID] = struct{}{}

		// Если PID не было в предыдущем снапшоте — запрос новый
		if _, exists := d.previousPIDs[q.PID]; !exists {
			q.IsNew = true
		}
	}

	d.previousPIDs = currentPIDs
}

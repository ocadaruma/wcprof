package wcprof

import "time"

type Timer struct {
	ID string
	Start time.Time
	End time.Time
}

func NewTimer(id string) *Timer {
	return &Timer{
		ID: id,
		Start: time.Now(),
	}
}

func (timer *Timer) Stop() {
	timer.End = time.Now()
}

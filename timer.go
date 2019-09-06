package wcprof

import (
	"github.com/olekukonko/tablewriter"
	"io"
	"math"
	"strconv"
	"sync"
	"time"
)

type Registry struct {
	mu *sync.Mutex
	samples []*Timer
}

var defaultRegistry *Registry

func init() {
	defaultRegistry = &Registry{
		mu: &sync.Mutex{},
	}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func (registry *Registry) Write(w io.Writer) {
	result := registry.aggregate()
	writer := tablewriter.NewWriter(w)
	writer.SetHeader([]string{
		"Name", "count", "sum(ms)", "max(ms)", "min(ms)", "avg(ms)",
	})

	writer.SetAutoFormatHeaders(false)
	writer.SetAlignment(tablewriter.ALIGN_RIGHT)

	for id, row := range result.rows {
		writer.Append([]string{
			id,
			strconv.Itoa(row.count),
			formatDuration(row.sum),
			formatDuration(row.max),
			formatDuration(row.min),
			formatDuration(row.avg),
		})
	}
}

type Timer struct {
	ID string
	Start time.Time
	End time.Time
}

func NewTimer(id string) *Timer {
	timer := &Timer{
		ID:    id,
		Start: time.Now(),
	}
	return timer
}

func (timer *Timer) Stop() {
	timer.End = time.Now()

	defaultRegistry.mu.Lock()
	defaultRegistry.samples = append(defaultRegistry.samples, timer)
	defaultRegistry.mu.Unlock()
}

type resultRow struct {
	count int
	sum   time.Duration
	max   time.Duration
	min   time.Duration
	avg   time.Duration
}

type result struct {
	rows map[string]*resultRow
}

func (registry *Registry) aggregate() *result {
	rows := make(map[string]*resultRow)

	for _, timer := range registry.samples {
		duration := timer.End.Sub(timer.Start)

		row, ok := rows[timer.ID];
		if !ok {
			row = &resultRow{
				min: math.MaxInt64,
			}
		}

		row.count++
		row.sum += duration
		if duration < row.min {
			row.min = duration
		}
		if duration > row.max {
			row.max = duration
		}
		row.avg = row.sum / time.Duration(row.count)
	}
	return &result{rows: rows}
}

func formatDuration(d time.Duration) string {
	return strconv.FormatFloat(float64(d)/(1000*1000), 'f', -1, 64)
}

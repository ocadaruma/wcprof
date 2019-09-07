package wcprof

import (
	"github.com/olekukonko/tablewriter"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Registry struct {
	mu    *sync.Mutex
	stats map[string]*Stat
}

var defaultRegistry *Registry

var enabled bool

func init() {
	defaultRegistry = &Registry{
		mu:    &sync.Mutex{},
		stats: make(map[string]*Stat),
	}

	enabled = os.Getenv("WCPROF_OFF") == ""
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func Off() {
	enabled = false
}

func (registry *Registry) Write(w io.Writer) {
	result := registry.aggregate()
	writer := tablewriter.NewWriter(w)
	writer.SetHeader([]string{
		"Name", "count", "sum(ms)", "max(ms)", "min(ms)", "avg(ms)",
	})

	writer.SetAutoFormatHeaders(false)
	writer.SetAlignment(tablewriter.ALIGN_RIGHT)

	for _, row := range result.rows {
		writer.Append([]string{
			row.id,
			strconv.Itoa(row.count),
			formatDuration(row.sum),
			formatDuration(row.max),
			formatDuration(row.min),
			formatDuration(row.avg),
		})
	}

	writer.Render()
}

func (registry *Registry) Print() {
	registry.Write(os.Stdout)
}

type Timer struct {
	ID string
	Start time.Time
	End time.Time
}

type Stat struct {
	ID string
	Count int
	Sum   time.Duration
	Max   time.Duration
	Min   time.Duration
	Avg   time.Duration
}

func NewTimer(id string) *Timer {
	if !enabled {
		return nil
	}

	timer := &Timer{
		ID:    id,
		Start: time.Now(),
	}
	return timer
}

func (timer *Timer) Stop() {
	if !enabled {
		return
	}
	timer.End = time.Now()

	defaultRegistry.mu.Lock()
	stat, ok := defaultRegistry.stats[timer.ID]
	if !ok {
		stat = &Stat{
			ID:    timer.ID,
			Count: 0,
			Sum:   0,
			Max:   0,
			Min:   math.MaxInt64,
			Avg:   0,
		}
	}
	duration := timer.End.Sub(timer.Start)

	stat.Count++
	stat.Sum += duration

	if duration < stat.Min {
		stat.Min = duration
	}
	if duration > stat.Max {
		stat.Max = duration
	}
	stat.Avg = stat.Sum / time.Duration(stat.Count)
	defaultRegistry.stats[timer.ID] = stat

	defaultRegistry.mu.Unlock()
}

type resultRow struct {
	id    string
	count int
	sum   time.Duration
	max   time.Duration
	min   time.Duration
	avg   time.Duration
}

type result struct {
	rows []*resultRow
}

func (registry *Registry) aggregate() *result {
	var rows []*resultRow

	for id, stat := range registry.stats {
		rows = append(rows, &resultRow{
			id   : id,
			count: stat.Count,
			sum:   stat.Sum,
			max:   stat.Max,
			min:   stat.Min,
			avg:   stat.Avg,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].sum > rows[j].sum
	})

	return &result{rows: rows}
}

func formatDuration(d time.Duration) string {
	return strconv.FormatFloat(float64(d)/(1000*1000), 'f', -1, 64)
}

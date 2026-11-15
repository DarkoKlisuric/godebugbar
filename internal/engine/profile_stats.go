package engine

import "github.com/DarkoKlisuric/godebugbar/internal/contextstate"

type ProfileStats struct {
	Queries          int     `json:"queries"`
	QueryMs          float64 `json:"query_ms"`
	SlowQueries      int     `json:"slow_queries"`
	DuplicateQueries int     `json:"duplicate_queries"`
	NPlusOne         int     `json:"n_plus_one"`
	FailedQueries    int     `json:"failed_queries"`
	Errors           int     `json:"errors"`
	Logs             int     `json:"logs"`
	HTTPCalls        int     `json:"http_calls"`
	Panic            bool    `json:"panic,omitempty"`
}

func NewProfileStats(rec *contextstate.Recorder) ProfileStats {
	if rec == nil {
		return ProfileStats{}
	}

	dropped := rec.Dropped()
	analysis := contextstate.AnalyzeQueries(rec.Queries(), rec.Config())
	panicValue, _ := rec.Panic()

	return ProfileStats{
		Queries:          analysis.Count + dropped.Queries,
		QueryMs:          float64(analysis.Total.Microseconds()) / 1000.0,
		SlowQueries:      analysis.Slow,
		DuplicateQueries: analysis.Duplicates,
		NPlusOne:         analysis.NPlusOne,
		FailedQueries:    analysis.Failed,
		Errors:           len(rec.Errors()) + dropped.Errors,
		Logs:             len(rec.Logs()) + dropped.Logs,
		HTTPCalls:        len(rec.HTTPCalls()) + dropped.HTTPCalls,
		Panic:            panicValue != "",
	}
}

package contextstate

import (
	"strings"
	"time"
)

type QueryAnalysis struct {
	Count      int
	Total      time.Duration
	Failed     int
	Slow       int
	Duplicates int
	NPlusOne   int
	byTemplate map[string]int
	bySQL      map[string]int
	config     Config
}

func AnalyzeQueries(queries []QueryEntry, config Config) *QueryAnalysis {
	a := &QueryAnalysis{
		Count:      len(queries),
		byTemplate: make(map[string]int, len(queries)),
		bySQL:      make(map[string]int, len(queries)),
		config:     config,
	}
	for _, q := range queries {
		a.Total += q.Duration
		a.byTemplate[TemplateOf(q)]++
		a.bySQL[identity(q)]++
		if q.Error != "" {
			a.Failed++
		}
		if a.IsSlow(q) {
			a.Slow++
		}
	}
	for _, count := range a.bySQL {
		if count > 1 {
			a.Duplicates += count - 1
		}
	}
	for _, count := range a.byTemplate {
		if count >= config.NPlusOneThreshold {
			a.NPlusOne++
		}
	}

	return a
}

func (a *QueryAnalysis) IsSlow(q QueryEntry) bool {
	return q.Duration >= a.config.SlowQuery
}

func (a *QueryAnalysis) IsDuplicate(q QueryEntry) bool {
	return a.bySQL[identity(q)] > 1
}

func (a *QueryAnalysis) Repeats(q QueryEntry) int {
	return a.byTemplate[TemplateOf(q)]
}

func (a *QueryAnalysis) IsNPlusOne(q QueryEntry) bool {
	return a.Repeats(q) >= a.config.NPlusOneThreshold
}

func (a *QueryAnalysis) RepeatedTemplates() map[string]int {
	out := make(map[string]int)
	for template, count := range a.byTemplate {
		if count >= a.config.NPlusOneThreshold {
			out[template] = count
		}
	}

	return out
}

func TemplateOf(q QueryEntry) string {
	if q.Template != "" {
		return q.Template
	}

	return q.SQL
}

func identity(q QueryEntry) string {
	return q.SQL + "\x00" + strings.Join(q.Vars, "\x00")
}

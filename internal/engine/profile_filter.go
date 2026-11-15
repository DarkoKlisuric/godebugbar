package engine

import (
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const (
	FlagError     = "error"
	FlagSlow      = "slow"
	FlagNPlusOne  = "n_plus_one"
	FlagDuplicate = "duplicate"
	FlagPanic     = "panic"
)

type ProfileFilter struct {
	Since       string
	Text        string
	Method      string
	StatusClass int
	Status      int
	MinMs       float64
	Flags       []string
	Limit       int
}

func ParseProfileFilter(values url.Values) ProfileFilter {
	f := ProfileFilter{
		Since:  values.Get("since"),
		Text:   strings.ToLower(strings.TrimSpace(values.Get("q"))),
		Method: strings.ToUpper(strings.TrimSpace(values.Get("method"))),
	}
	if status := strings.TrimSpace(values.Get("status")); status != "" {
		if n, err := strconv.Atoi(strings.TrimRight(strings.ToLower(status), "x")); err == nil {
			if n < 10 {
				f.StatusClass = n
			} else {
				f.Status = n
			}
		}
	}
	if minMs, err := strconv.ParseFloat(values.Get("min_ms"), 64); err == nil {
		f.MinMs = minMs
	}
	if limit, err := strconv.Atoi(values.Get("limit")); err == nil && limit > 0 {
		f.Limit = limit
	}
	for _, flag := range strings.Split(values.Get("has"), ",") {
		if flag = strings.TrimSpace(flag); flag != "" {
			f.Flags = append(f.Flags, flag)
		}
	}

	return f
}

func (f *ProfileFilter) Matches(p *Profile) bool {
	switch {
	case f.Method != "" && p.Method != f.Method:
		return false
	case f.StatusClass != 0 && p.Status/100 != f.StatusClass:
		return false
	case f.Status != 0 && p.Status != f.Status:
		return false
	case p.DurationMs < f.MinMs:
		return false
	}
	for _, flag := range f.Flags {
		if !hasFlag(p, flag) {
			return false
		}
	}

	return f.Text == "" || matchesText(p, f.Text)
}

func hasFlag(p *Profile, flag string) bool {
	switch flag {
	case FlagError:
		return p.Status >= 500 || p.Stats.Errors > 0 || p.Stats.FailedQueries > 0 || p.Stats.Panic
	case FlagSlow:
		return p.Stats.SlowQueries > 0
	case FlagNPlusOne:
		return p.Stats.NPlusOne > 0
	case FlagDuplicate:
		return p.Stats.DuplicateQueries > 0
	case FlagPanic:
		return p.Stats.Panic
	default:
		return true
	}
}

func matchesText(p *Profile, text string) bool {
	fields := []string{p.ID, p.Method, p.Path, p.Route, strconv.Itoa(p.Status)}
	fields = append(fields, p.Errors...)
	for _, q := range p.Queries {
		fields = append(fields, q.SQL, q.Template)
	}

	return slices.ContainsFunc(fields, func(field string) bool {
		return strings.Contains(strings.ToLower(field), text)
	})
}

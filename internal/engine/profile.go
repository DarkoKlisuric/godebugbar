package engine

import (
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

type Profile struct {
	ID          string                    `json:"id"`
	Method      string                    `json:"method"`
	Path        string                    `json:"path"`
	Route       string                    `json:"route,omitempty"`
	Status      int                       `json:"status"`
	Size        int                       `json:"size"`
	DurationMs  float64                   `json:"duration_ms"`
	CreatedAt   time.Time                 `json:"created_at"`
	Explainable bool                      `json:"explainable,omitempty"`
	Stats       ProfileStats              `json:"stats"`
	Sections    []contract.Section        `json:"sections"`
	Queries     []contextstate.QueryEntry `json:"-"`
	Errors      []string                  `json:"-"`
}

type ProfileSummary struct {
	ID         string                 `json:"id"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Route      string                 `json:"route,omitempty"`
	Status     int                    `json:"status"`
	DurationMs float64                `json:"duration_ms"`
	CreatedAt  time.Time              `json:"created_at"`
	Stats      ProfileStats           `json:"stats"`
	Badges     []contract.ToolbarItem `json:"badges"`
}

func (p *Profile) Summary() ProfileSummary {
	badges := make([]contract.ToolbarItem, 0, len(p.Sections))
	for _, sec := range p.Sections {
		badges = append(badges, sec.Toolbar)
	}

	return ProfileSummary{
		ID:         p.ID,
		Method:     p.Method,
		Path:       p.Path,
		Route:      p.Route,
		Status:     p.Status,
		DurationMs: p.DurationMs,
		CreatedAt:  p.CreatedAt,
		Stats:      p.Stats,
		Badges:     badges,
	}
}

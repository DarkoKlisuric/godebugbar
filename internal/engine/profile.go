package engine

import (
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
)

type Profile struct {
	ID         string             `json:"id"`
	Method     string             `json:"method"`
	Path       string             `json:"path"`
	Route      string             `json:"route,omitempty"`
	Status     int                `json:"status"`
	Size       int                `json:"size"`
	DurationMs float64            `json:"duration_ms"`
	CreatedAt  time.Time          `json:"created_at"`
	Sections   []contract.Section `json:"sections"`
}

package contract

type TimelineEvent struct {
	Label      string     `json:"label"`
	Category   string     `json:"category"`
	StartMs    float64    `json:"start_ms"`
	DurationMs float64    `json:"duration_ms"`
	Color      BadgeColor `json:"color,omitempty"`
}

package contract

import (
	"context"
	"net/http"
	"time"
)

type BadgeColor string

const (
	ColorDefault BadgeColor = "default"
	ColorGreen   BadgeColor = "green"
	ColorYellow  BadgeColor = "yellow"
	ColorRed     BadgeColor = "red"
)

func StatusColor(status int) BadgeColor {
	switch {
	case status >= http.StatusInternalServerError:
		return ColorRed
	case status >= http.StatusBadRequest:
		return ColorYellow
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return ColorGreen
	default:
		return ColorDefault
	}
}

type ToolbarItem struct {
	Label   string            `json:"label"`
	Value   string            `json:"value"`
	Color   BadgeColor        `json:"color"`
	Tooltip map[string]string `json:"tooltip,omitempty"`
}

type ResponseInfo struct {
	Status   int
	Size     int
	Duration time.Duration
	Headers  http.Header
	Route    string
	Errors   []string
}

type Section struct {
	Name    string      `json:"name"`
	Toolbar ToolbarItem `json:"toolbar"`
	Panel   *Panel      `json:"panel,omitempty"`
}

func NewSection(toolbar ToolbarItem, panel *Panel) *Section {
	return &Section{
		Toolbar: toolbar,
		Panel:   panel,
	}
}

type Collector interface {
	Name() string
	Collect(ctx context.Context, request *http.Request, response *ResponseInfo) *Section
}

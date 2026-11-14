package contract

import (
	"net/http"
	"testing"
)

func TestStatusColorAndSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   BadgeColor
	}{
		{name: "100 informational is default", status: http.StatusContinue, want: ColorDefault},
		{name: "200 ok is green", status: http.StatusOK, want: ColorGreen},
		{name: "404 not found is yellow", status: http.StatusNotFound, want: ColorYellow},
		{name: "500 internal error is red", status: http.StatusInternalServerError, want: ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := StatusColor(tt.status); got != tt.want {
				t.Errorf("StatusColor(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}

	panel := NewPanel("Auth")
	section := NewSection(ToolbarItem{Label: "User", Value: "admin", Color: ColorGreen}, panel)
	if section.Toolbar.Label != "User" || section.Panel != panel {
		t.Errorf("section = %+v, want User and panel", section)
	}
}

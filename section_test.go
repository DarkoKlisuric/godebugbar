package godebugbar

import "testing"

func TestNewSection(t *testing.T) {
	t.Parallel()

	panel := NewPanel("Auth")
	section := NewSection(ToolbarItem{Label: "User", Value: "admin", Color: ColorGreen}, panel)

	if section.Toolbar.Label != "User" || section.Toolbar.Value != "admin" {
		t.Errorf("section.Toolbar = %+v, want User=admin", section.Toolbar)
	}

	if section.Panel != panel {
		t.Errorf("section.Panel = %v, want %v", section.Panel, panel)
	}
}

package ui_test

import (
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/ui"
)

func TestToolbarScriptAndDashboard(t *testing.T) {
	t.Parallel()

	script := ui.ToolbarScript()
	if !strings.Contains(script, "GoDebugBar") {
		t.Fatalf("expected embedded toolbar script to contain GoDebugBar")
	}

	page := ui.DashboardHTML("/_custom")
	if !strings.Contains(page, "/_custom/assets/toolbar.js") {
		t.Fatalf("expected dashboard HTML to contain base URL")
	}
}

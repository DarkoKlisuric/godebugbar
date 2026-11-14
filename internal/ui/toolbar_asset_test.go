package ui_test

import (
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/ui"
)

func TestToolbarScript(t *testing.T) {
	t.Parallel()

	script := ui.ToolbarScript()
	if !strings.Contains(script, "GoDebugBar") {
		t.Fatalf("expected embedded toolbar script to contain GoDebugBar")
	}
}

package godebugbar

import (
	"strings"
	"testing"
)

func TestToolbarAsset(t *testing.T) {
	t.Parallel()

	if !strings.Contains(toolbarScript, "GoDebugbar") {
		t.Error("toolbarScript does not contain GoDebugbar")
	}
}

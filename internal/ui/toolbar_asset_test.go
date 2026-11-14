package ui_test

import (
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/ui"
)

func TestAssets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		asset string
		want  []string
		not   []string
	}{
		{name: "render script", asset: ui.RenderScript(), want: []string{"window.GoDebugBarRender"}},
		{name: "toolbar bundles the renderer first", asset: ui.ToolbarScript(), want: []string{"window.GoDebugBarRender", "godebugbar-host", "XMLHttpRequest.prototype.open"}},
		{
			name:  "dashboard uses the base URL",
			asset: ui.DashboardHTML("/_custom"),
			want:  []string{`src="/_custom/assets/render.js"`, "const baseUrl = '/_custom';"},
			not:   []string{"{{BASE_URL", "toolbar.js"},
		},
		{
			name:  "dashboard escapes the base URL",
			asset: ui.DashboardHTML(`/x'"y`),
			want:  []string{`src="/x&#39;&#34;y/assets/render.js"`, `const baseUrl = '/x\'\"y';`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, want := range tc.want {
				if !strings.Contains(tc.asset, want) {
					t.Errorf("expected %q", want)
				}
			}
			for _, not := range tc.not {
				if strings.Contains(tc.asset, not) {
					t.Errorf("did not expect %q", not)
				}
			}
		})
	}

	if strings.Index(ui.ToolbarScript(), "GoDebugBarRender =") > strings.Index(ui.ToolbarScript(), "godebugbar-host") {
		t.Fatalf("the renderer must come before the toolbar")
	}
}

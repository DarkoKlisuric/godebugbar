package ui

import (
	_ "embed"
	"strings"
)

//go:embed assets/toolbar.js
var toolbarScript string

//go:embed assets/dashboard.html
var dashboardHTML string

func ToolbarScript() string {
	return toolbarScript
}

func DashboardHTML(baseURL string) string {
	return strings.ReplaceAll(dashboardHTML, "{{BASE_URL}}", baseURL)
}

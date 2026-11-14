package ui

import (
	_ "embed"
	"html"
	"strings"
	"text/template"
)

const (
	baseURLPlaceholder   = "{{BASE_URL}}"
	baseURLJSPlaceholder = "{{BASE_URL_JS}}"
)

//go:embed assets/render.js
var renderScript string

//go:embed assets/toolbar.js
var toolbarScript string

//go:embed assets/dashboard.html
var dashboardHTML string

func RenderScript() string {
	return renderScript
}

func ToolbarScript() string {
	return renderScript + "\n" + toolbarScript
}

func DashboardHTML(baseURL string) string {
	return strings.NewReplacer(
		baseURLJSPlaceholder, template.JSEscapeString(baseURL),
		baseURLPlaceholder, html.EscapeString(baseURL),
	).Replace(dashboardHTML)
}

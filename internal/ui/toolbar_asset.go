package ui

import _ "embed"

//go:embed assets/toolbar.js
var toolbarScript string

func ToolbarScript() string {
	return toolbarScript
}

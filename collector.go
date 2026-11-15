package godebugbar

import "github.com/DarkoKlisuric/godebugbar/internal/contract"

// Contracts for custom collectors.
type (
	// Collector turns a finished request into a toolbar badge and an inspector
	// tab. Returning nil leaves the section out.
	Collector = contract.Collector
	// Section is a toolbar badge plus a panel.
	Section = contract.Section
	// ToolbarItem is a toolbar badge.
	ToolbarItem = contract.ToolbarItem
	// ResponseInfo describes the finished response.
	ResponseInfo = contract.ResponseInfo
	// BadgeColor colors a badge, a table row or a timeline bar.
	BadgeColor = contract.BadgeColor
)

// Badge colors.
const (
	ColorDefault = contract.ColorDefault
	ColorGreen   = contract.ColorGreen
	ColorYellow  = contract.ColorYellow
	ColorRed     = contract.ColorRed
)

// StatusColor maps an HTTP status code to a badge color.
func StatusColor(status int) BadgeColor {
	return contract.StatusColor(status)
}

// NewSection builds a Section.
func NewSection(item ToolbarItem, panel *Panel) *Section {
	return contract.NewSection(item, panel)
}

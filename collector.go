package godebugbar

import "github.com/DarkoKlisuric/godebugbar/internal/contract"

type (
	Collector    = contract.Collector
	Section      = contract.Section
	ToolbarItem  = contract.ToolbarItem
	ResponseInfo = contract.ResponseInfo
	BadgeColor   = contract.BadgeColor
)

const (
	ColorDefault = contract.ColorDefault
	ColorGreen   = contract.ColorGreen
	ColorYellow  = contract.ColorYellow
	ColorRed     = contract.ColorRed
)

func StatusColor(status int) BadgeColor {
	return contract.StatusColor(status)
}

func NewSection(item ToolbarItem, panel *Panel) *Section {
	return contract.NewSection(item, panel)
}

package godebugbar

type Section struct {
	Name    string      `json:"name"`
	Toolbar ToolbarItem `json:"toolbar"`
	Panel   *Panel      `json:"panel,omitempty"`
}

func NewSection(toolbar ToolbarItem, panel *Panel) *Section {
	return &Section{
		Toolbar: toolbar,
		Panel:   panel,
	}
}

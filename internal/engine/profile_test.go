package engine_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DarkoKlisuric/godebugbar/internal/contract"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestProfileJSON(t *testing.T) {
	t.Parallel()

	sec := contract.NewSection(contract.ToolbarItem{
		Label: "200",
		Value: "12.5ms",
		Color: contract.ColorGreen,
	}, contract.NewPanel("Request"))
	sec.Name = "request"

	p := engine.Profile{
		ID:         "prof-1",
		Method:     "GET",
		Path:       "/orders",
		Status:     200,
		DurationMs: 12.5,
		CreatedAt:  time.Unix(1700000000, 0).UTC(),
		Sections:   []contract.Section{*sec},
	}

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded engine.Profile
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded.ID != "prof-1" || decoded.Status != 200 || len(decoded.Sections) != 1 {
		t.Fatalf("unexpected decoded profile: %+v", decoded)
	}
}

package godebugbar

import (
	"testing"
	"time"
)

func TestProfile(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	profile := &Profile{
		ID:         "req-1",
		Method:     "GET",
		Path:       "/v1/plans",
		Route:      "/v1/plans",
		Status:     200,
		DurationMs: 4.5,
		Size:       128,
		StartedAt:  now,
	}

	if profile.ID != "req-1" || profile.Status != 200 {
		t.Errorf("profile = %+v, want ID=req-1 and Status=200", profile)
	}
}

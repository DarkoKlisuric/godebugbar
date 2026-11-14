package gormadapter

import (
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestGormPlugin(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("gorm.Open(DummyDialector) error = %v", err)
	}

	godebugbar.New(
		Attach(db),
	)

	recorder := godebugbar.NewRecorder()
	ctx := godebugbar.WithRecorder(t.Context(), recorder)

	var row struct {
		ID string
	}
	stmt := db.WithContext(ctx).Table("billing.plans").Where("id = ?", "pro").Find(&row)

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("len(queries) = %d, want 1", len(queries))
	}

	if !strings.Contains(queries[0].SQL, "SELECT * FROM `billing`.`plans`") {
		t.Errorf("SQL = %q, want SELECT * FROM `billing`.`plans`", queries[0].SQL)
	}

	if len(queries[0].Vars) != 1 || queries[0].Vars[0] != "pro" {
		t.Errorf("Vars = %v, want [pro]", queries[0].Vars)
	}

	stmt.Error = errors.New("connection reset")
	plugin := NewGormPlugin()
	plugin.after(stmt)

	queries = recorder.Queries()
	if len(queries) != 2 || queries[1].Error != "connection reset" {
		t.Errorf("Error query = %+v, want Error=connection reset", queries)
	}
}

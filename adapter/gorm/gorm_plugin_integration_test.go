//go:build integration

package gormadapter

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/DarkoKlisuric/godebugbar"
)

func TestGormPluginPostgresIntegration(t *testing.T) {
	t.Parallel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Discard,
	})
	if err != nil {
		t.Fatalf("gorm.Open(postgres) error = %v", err)
	}

	godebugbar.New(
		Attach(db),
	)

	recorder := godebugbar.NewRecorder()
	ctx := godebugbar.WithRecorder(t.Context(), recorder)

	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		t.Fatalf("Raw(SELECT 1) error = %v", err)
	}

	if result != 1 {
		t.Errorf("result = %d, want 1", result)
	}

	queries := recorder.Queries()
	if len(queries) != 1 || queries[0].SQL != "SELECT 1" {
		t.Errorf("recorded queries = %+v, want [SELECT 1]", queries)
	}
}

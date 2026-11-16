package pgxadapter_test

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DarkoKlisuric/godebugbar"
	pgxadapter "github.com/DarkoKlisuric/godebugbar/adapter/pgx"
)

func ExampleNewTracer() {
	config, err := pgxpool.ParseConfig("postgres://localhost:5432/app")
	if err != nil {
		fmt.Println(err)
		return
	}
	config.ConnConfig.Tracer = pgxadapter.NewTracer()

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		pgxadapter.Attach(),
	)

	fmt.Println(bar != nil, config.ConnConfig.Tracer != nil)
	// Output: true true
}

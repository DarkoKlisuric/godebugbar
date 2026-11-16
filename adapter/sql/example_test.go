package sqladapter_test

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DarkoKlisuric/godebugbar"
	sqladapter "github.com/DarkoKlisuric/godebugbar/adapter/sql"
)

func ExampleOpen() {
	db, err := sqladapter.Open("pgx", "postgres://localhost:5432/app")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = db.Close() }()

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		sqladapter.Attach(),
	)

	fmt.Println(bar != nil, db.Driver() != nil)
	// Output: true true
}

package gormadapter_test

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/DarkoKlisuric/godebugbar"
	gormadapter "github.com/DarkoKlisuric/godebugbar/adapter/gorm"
)

func ExampleAttach() {
	db, err := gorm.Open(postgres.Open("postgres://localhost:5432/app"), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		fmt.Println(err)
		return
	}

	bar := godebugbar.New(
		godebugbar.WithEnabled(true),
		gormadapter.Attach(db),
	)

	fmt.Println(bar != nil, db.Plugins["godebugbar"] != nil)
	// Output: true true
}

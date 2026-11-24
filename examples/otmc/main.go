/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	_ "embed"

	"github.com/otmc-sw/rest/examples/otmc/models"
	"github.com/otmc-sw/rest/otmc"
)

//go:embed configs/banner.txt
var APP_BANNER string

func main() {
	app := otmc.New(otmc.Config{
		Name:   "My Application",
		Port:   6000,
		Banner: APP_BANNER,
	})

	app.Register(
		&models.Role{},
		&models.User{},
	)

	app.Runner()
}

/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	_ "embed"

	"github.com/otmc-sw/rest/examples/app/handlers"
	"github.com/otmc-sw/rest/examples/app/models"
	"github.com/otmc-sw/rest/otmc"
)

func main() {
	app := otmc.New(otmc.Config{
		Name:        "My Application",
		Port:        6000,
		Banner:      "configs/banner.txt",
		InitialData: "db/*.sql",
	})

	app.Register(
		&models.Role{},
		&models.User{},
	)

	// Register custom handlers
	app.RegisterHandlers(handlers.RegisterUserHandlers)

	app.Runner()
}

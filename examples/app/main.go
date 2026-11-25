/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	_ "embed"

	"github.com/otmc-sw/rest/app"
	"github.com/otmc-sw/rest/examples/app/handlers"
	"github.com/otmc-sw/rest/examples/app/models"
)

func main() {
	myapp := app.New(app.Config{
		Name:        "My Application",
		Port:        6000,
		Banner:      "configs/banner.txt",
		InitialData: "db/*.sql",
	})

	myapp.Register(
		&models.Role{},
		&models.User{},
	)

	myapp.RegisterHandlers(handlers.RegisterUserHandlers)
	myapp.Runner()
}

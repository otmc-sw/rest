/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/otmc-sw/rest/examples/otmc/models"
	"github.com/otmc-sw/rest/otmc"
)

func main() {
	port := 5007
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	app := otmc.New(otmc.Config{
		Name: "My Application",
		Port: port,

		Database: otmc.DatabaseConfig{
			Driver: "sqlite",
			DSN:    "./data/app.db",
		},
	})

	app.Register(
		&models.Role{},
		&models.User{},
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

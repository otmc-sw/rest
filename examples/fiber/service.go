/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	"fmt"

	rest "github.com/otmc-sw/rest"
)

// User is the domain entity owned by the service layer.
type User struct {
	ID    string
	Name  string
	Email string
}

// UserService owns the business logic.
type UserService struct{}

// userService is the shared instance used by the handlers.
var userService = &UserService{}

// Create loads, validates business rules, persists, and returns the entity.
func (s *UserService) Create(ctx rest.Context, req CreateUserRequest) (User, error) {
	// In a real service you would validate business rules and persist.
	// Here we synthesize an ID for demonstration.
	user := User{
		ID:    fmt.Sprintf("usr_%s", req.Email),
		Name:  req.Name,
		Email: req.Email,
	}
	return user, nil
}
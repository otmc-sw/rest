/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package main

import (
	rest "github.com/otmc-sw/rest"
)

// CreateUserRequest is the inbound DTO bound from the request body.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserResponse is the outbound DTO written in the response.
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUser describes only the business flow. The pipeline orchestrates
// Request -> Bind -> Validate -> Business -> Map -> Respond.
func CreateUser(c FiberContext) error {
	return rest.
		Create[CreateUserRequest, User, UserResponse](c).
		Bind().
		Validate(func(r CreateUserRequest) error {
			return rest.Validate().Required(r.Name).Email(r.Email).Validate()
		}).
		Handle(userService.Create).
		Respond()
}
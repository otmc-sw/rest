/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	rest "github.com/otmc-sw/rest"
	"github.com/otmc-sw/rest/examples/fiber/services"
)

var userService = services.NewUserService()

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateUserHandler(ctx rest.Context, req services.CreateUserRequest) (services.User, error) {
	return userService.Create(ctx.Context(), req)
}

func CreateUser(c rest.Context) error {
	return rest.
		Create[services.CreateUserRequest, services.User, UserResponse](c).
		Bind().
		Validate(func(r services.CreateUserRequest) error {
			return rest.Validate().Required(r.Name).Email(r.Email).Validate()
		}).
		Handle(CreateUserHandler).
		Respond()
}

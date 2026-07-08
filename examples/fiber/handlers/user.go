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

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func CreateUser(c rest.Context) error {
	return rest.
		Create[services.CreateUserRequest, services.User, UserResponse](c).
		Bind().
		Validate(services.Validates()).
		Handle(services.CreateUserHandler()).
		Respond()
}

func GetUser(c rest.Context) error {
	return rest.
		Get[struct{}, services.User, UserResponse](c).
		Handle(services.GetUserHandler()).
		Respond()
}

func UpdateUser(c rest.Context) error {
	return rest.
		Update[services.UpdateUserRequest, services.User, UserResponse](c).
		Bind().
		Validate(services.ValidatesUpdate()).
		Handle(services.UpdateUserHandler()).
		Respond()
}

func DeleteUser(c rest.Context) error {
	return rest.
		Delete[UserResponse](c).
		Handle(services.DeleteUserHandler()).
		Respond()
}

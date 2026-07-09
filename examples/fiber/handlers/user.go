/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package handlers

import (
	"strconv"

	rest "github.com/otmc-sw/rest"
	sqlc "github.com/otmc-sw/rest/examples/fiber/db/sqlc"
)

var database *sqlc.Queries

type User struct {
	ID    string
	Name  string
	Email string
}

type UserRequest struct {
	Name  string
	Email string
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ValidateUser(r UserRequest) error {
	return rest.Validate().
		Required(r.Name).
		Email(r.Email).
		Validate()
}

func CreateUser() error {
	return rest.
		Create[UserRequest, User, UserResponse](FiberContext{}).
		Bind().
		Validate(ValidateUser).
		Handle(func(ctx rest.Context, req UserRequest) (User, error) {
			params := sqlc.CreateUserParams{
				Username: req.Email,
				Email:    req.Email,
			}
			err := database.CreateUser(ctx.Context(), params)
			if err != nil {
				return User{}, err
			}
			return User{
				ID:    req.Email,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		}).
		Respond()
}

func GetUser() error {
	return rest.
		Get[struct{}, User, UserResponse](FiberContext{}).
		Handle(func(ctx rest.Context, req struct{}) (User, error) {
			id := c.Param("id")
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return User{}, err
			}
			user, err := database.GetUser(ctx.Context(), intID)
			if err != nil {
				return User{}, err
			}
			return User{
				ID:    strconv.FormatInt(user.ID, 10),
				Name:  user.Username,
				Email: user.Email,
			}, nil
		}).
		Respond()
}

func UpdateUser() error {
	return rest.
		Update[UserRequest, User, UserResponse](FiberContext{}).
		Bind().
		Validate(ValidateUser).
		Handle(func(ctx rest.Context, req UserRequest) (User, error) {
			id := c.Param("id")
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return User{}, err
			}
			params := sqlc.UpdateUserParams{
				Username: req.Email,
				Email:    req.Email,
				ID:       intID,
			}
			err = database.UpdateUser(ctx.Context(), params)
			if err != nil {
				return User{}, err
			}
			return User{
				ID:    strconv.FormatInt(intID, 10),
				Name:  req.Name,
				Email: req.Email,
			}, nil
		}).
		Respond()
}

func DeleteUser() error {
	return rest.
		Delete[UserResponse](FiberContext{}).
		Handle(func(ctx rest.Context, req struct{}) (struct{}, error) {
			id := c.Param("id")
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return struct{}{}, err
			}
			return struct{}{}, database.DeleteUser(ctx.Context(), intID)
		}).
		Respond()
}

/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package services

import (
	"context"
	"fmt"

	"github.com/otmc-sw/rest"
)

type User struct {
	ID    string
	Name  string
	Email string
}

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

var userService = NewUserService()

type CreateUserRequest struct {
	Name  string
	Email string
}

type UpdateUserRequest struct {
	Name  string
	Email string
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (User, error) {
	user := User{
		ID:    fmt.Sprintf("usr_%s", req.Email),
		Name:  req.Name,
		Email: req.Email,
	}
	return user, nil
}

func (s *UserService) Get(ctx context.Context, id string) (User, error) {
	user := User{
		ID:    id,
		Name:  "John Doe",
		Email: "john@example.com",
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (User, error) {
	user := User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return nil
}

func Validates() func(r CreateUserRequest) error {
	return func(r CreateUserRequest) error {
		return rest.Validate().
			Required(r.Name).
			Email(r.Email).
			Validate()
	}
}

func ValidatesUpdate() func(r UpdateUserRequest) error {
	return func(r UpdateUserRequest) error {
		return rest.Validate().
			Required(r.Name).
			Email(r.Email).
			Validate()
	}
}

func CreateUserHandler() func(ctx rest.Context, req CreateUserRequest) (User, error) {
	return func(ctx rest.Context, req CreateUserRequest) (User, error) {
		return userService.Create(ctx.Context(), req)
	}
}

func GetUserHandler() func(ctx rest.Context, req struct{}) (User, error) {
	return func(ctx rest.Context, req struct{}) (User, error) {
		id := ctx.Param("id")
		return userService.Get(ctx.Context(), id)
	}
}

func UpdateUserHandler() func(ctx rest.Context, req UpdateUserRequest) (User, error) {
	return func(ctx rest.Context, req UpdateUserRequest) (User, error) {
		id := ctx.Param("id")
		return userService.Update(ctx.Context(), id, req)
	}
}

func DeleteUserHandler() func(ctx rest.Context, req struct{}) (struct{}, error) {
	return func(ctx rest.Context, req struct{}) (struct{}, error) {
		id := ctx.Param("id")
		err := userService.Delete(ctx.Context(), id)
		return struct{}{}, err
	}
}

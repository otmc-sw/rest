/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/
package services

import (
	"context"
	"fmt"
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

type CreateUserRequest struct {
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

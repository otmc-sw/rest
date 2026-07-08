/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package response

import (
	"github.com/otmc-sw/rest/errors"
)

// SuccessResponse represents standard REST success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// Builder provides fluent API for building responses
type Builder struct {
	statusCode int
	data       interface{}
	message    string
	errBuilder *errors.Builder
}

// OK creates a 200 OK response builder
func OK() *Builder {
	return &Builder{
		statusCode: 200,
	}
}

// Created creates a 201 Created response builder
func Created() *Builder {
	return &Builder{
		statusCode: 201,
	}
}

// Accepted creates a 202 Accepted response builder
func Accepted() *Builder {
	return &Builder{
		statusCode: 202,
	}
}

// NoContent creates a 204 No Content response builder
func NoContent() *Builder {
	return &Builder{
		statusCode: 204,
	}
}

// Error creates an error response builder
func Error() *errors.Builder {
	return errors.New()
}

// Data sets the response data
func (b *Builder) Data(data interface{}) *Builder {
	b.data = data
	return b
}

// Message sets the response message
func (b *Builder) Message(msg string) *Builder {
	b.message = msg
	return b
}

// Build finalizes the success response
func (b *Builder) Build() SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    b.data,
		Message: b.message,
	}
}

// StatusCode returns the HTTP status code
func (b *Builder) StatusCode() int {
	return b.statusCode
}

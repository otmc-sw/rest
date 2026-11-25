/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"

	"github.com/otmc-sw/rest/errors"
	"github.com/otmc-sw/rest/otmc"
	"github.com/otmc-sw/rest/otmc/httpapi"
	"github.com/otmc-sw/rest/otmc/store"
	"github.com/otmc-sw/rest/response"
)

// RegisterUserHandlers registers custom user-related handlers.
func RegisterUserHandlers(app *fiber.App, prefix string, mgr *store.Manager) {
	userRepo := httpapi.GetRepository(mgr, "User")
	if userRepo == nil {
		return
	}

	base := prefix + "/user"

	// Custom user routes
	app.Post(base+"/login", Login(userRepo))
	app.Post(base+"/logout", Logout)
	app.Post(base+"/:id/avatar", UploadAvatar(userRepo))
}

// Login handles user authentication.
func Login(repo *store.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := httpapi.FiberContext{Ctx: c}

		// Parse request body
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return errors.New().BadRequest().Summary("invalid request body").Send(ctx)
		}

		// Simple authentication logic (in production, use proper password hashing)
		// This is a simplified example
		query := store.Query{
			Filters: map[string]string{
				"username": req.Username,
			},
		}

		result, err := repo.List(c.Context(), query)
		if err != nil {
			return errors.New().InternalError().Summary("authentication failed").Send(ctx)
		}

		if len(result.Items) == 0 {
			return errors.New().Unauthorized().Summary("invalid credentials").Send(ctx)
		}

		// In production, verify password hash here
		// For this example, we'll just return success
		user := result.Items[0]

		// Extract ID from user struct
		userID := getUserID(user)

		return response.OK[map[string]any](ctx).Data(map[string]any{
			"user":  user,
			"token": "fake-jwt-token-" + fmt.Sprint(userID),
		}).Message("login successful").Send()
	}
}

// Logout handles user logout.
func Logout(c *fiber.Ctx) error {
	ctx := httpapi.FiberContext{Ctx: c}
	return response.OK[string](ctx).Message("logout successful").Send()
}

// UploadAvatar handles avatar upload for a user.
func UploadAvatar(repo *store.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := httpapi.FiberContext{Ctx: c}

		// Parse user ID
		id, err := repo.ParseID(c.Params("id"))
		if err != nil {
			return errors.New().BadRequest().Summary("invalid user id").Send(ctx)
		}

		// Get the file from form
		file, err := c.FormFile("avatar")
		if err != nil {
			return errors.New().BadRequest().Summary("avatar file is required").Send(ctx)
		}

		// In production, upload to cloud storage and get URL
		// For this example, we'll just simulate it
		avatarURL := fmt.Sprintf("/uploads/avatars/%s-%s", c.Params("id"), file.Filename)

		// Update user with avatar URL - need to pass JSON bytes
		updateData := map[string]any{
			"avatar_url": avatarURL,
		}
		body, err := json.Marshal(updateData)
		if err != nil {
			return errors.New().InternalError().Summary("failed to marshal update data").Send(ctx)
		}

		updatedUser, err := repo.Patch(c.Context(), id, body)
		if err != nil {
			return errors.New().InternalError().Summary("failed to update avatar").Send(ctx)
		}

		return response.OK[any](ctx).Data(updatedUser).Message("avatar uploaded successfully").Send()
	}
}

// getUserID extracts the ID field from a user struct using reflection
func getUserID(user any) any {
	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		idField := v.FieldByName("ID")
		if idField.IsValid() {
			return idField.Interface()
		}
	}
	return nil
}

// HandlerFunc is the type signature for custom handlers.
// This matches otmc.HandlerFunc for easy registration.
type HandlerFunc = otmc.HandlerFunc

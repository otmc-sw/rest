/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/rest/examples/fiber/db"
	"github.com/otmc-sw/rest/examples/fiber/handlers"
)

type testResponse struct {
	Success  bool            `json:"success"`
	Data     json.RawMessage `json:"data,omitempty"`
	Error    *testError      `json:"error,omitempty"`
	Metadata *testMeta       `json:"_metadata,omitempty"`
}

type testError struct {
	Code    int         `json:"code"`
	Key     string      `json:"key"`
	Summary string      `json:"summary"`
	Detail  string      `json:"detail"`
}

type testMeta struct {
	Total   int `json:"total,omitempty"`
	Limit   int `json:"limit,omitempty"`
	Offset  int `json:"offset,omitempty"`
}

type userResponse struct {
	ID       int64           `json:"id"`
	Username string          `json:"username"`
	FullName string          `json:"full_name,omitempty"`
	Email    string          `json:"email"`
	Content  json.RawMessage `json:"content,omitempty"`
}

func setupTestApp(t *testing.T) (*fiber.App, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "rest-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.OpenDatabase(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := database.MigrateSchemas(); err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate test database: %v", err)
	}

	handlers.New(database.Queries)

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	app.Get("/test", handlers.TestResponse)
	app.Post("/users", handlers.CreateUser)
	app.Get("/users", handlers.GetAllUsers)
	app.Get("/users/:id", handlers.GetUser)
	app.Patch("/users/:id", handlers.UpdateUser)
	app.Delete("/users/:id", handlers.DeleteUser)

	cleanup := func() {
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return app, cleanup
}

func doRequest(app *fiber.App, method, path, body string, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return nil, err
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return app.Test(req, 5000)
}

func parseResponse(t *testing.T, resp *http.Response) testResponse {
	t.Helper()

	var tr testResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return tr
}

func parseUsers(t *testing.T, data json.RawMessage) []userResponse {
	t.Helper()

	var users []userResponse
	if err := json.Unmarshal(data, &users); err != nil {
		t.Fatalf("failed to decode users array: %v", err)
	}
	return users
}

func parseUser(t *testing.T, data json.RawMessage) userResponse {
	t.Helper()

	var user userResponse
	if err := json.Unmarshal(data, &user); err != nil {
		t.Fatalf("failed to decode user: %v", err)
	}
	return user
}


func TestTestEndpoint(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodGet, "/test", "", nil)
	if err != nil {
		t.Fatalf("GET /test failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}
}

func TestCreateUser_Success(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	payload := `{"username": "trung", "email": "trung@otmc.com.vn", "content": "{\"key\": \"value\"}"}`

	resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	if !tr.Success {
		t.Fatalf("expected success=true, got error: %+v", tr.Error)
	}

	user := parseUser(t, tr.Data)
	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if user.Username != "trung" {
		t.Fatalf("expected username 'trung', got '%s'", user.Username)
	}
	if user.Email != "trung@otmc.com.vn" {
		t.Fatalf("expected email 'trung@otmc.com.vn', got '%s'", user.Email)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	payload := `{"username": "duplicate", "email": "first@otmc.com.vn"}`

	resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
	if err != nil {
		t.Fatalf("first POST /users failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 on first create, got %d", resp.StatusCode)
	}

	payload2 := `{"username": "duplicate", "email": "second@otmc.com.vn"}`
	resp, err = doRequest(app, http.MethodPost, "/users", payload2, nil)
	if err != nil {
		t.Fatalf("second POST /users failed: %v", err)
	}
	if resp.StatusCode == http.StatusCreated {
		t.Fatal("expected duplicate creation to fail, but got 201")
	}
	t.Logf("duplicate create returned status %d (expected error)", resp.StatusCode)
}

func TestCreateUser_Validation_MissingUsername(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	payload := `{"email": "trung@otmc.dev"}`

	resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
	if err != nil {
		t.Fatalf("POST /users (missing username) failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing username, got %d", resp.StatusCode)
	}
}

func TestCreateUser_Validation_InvalidEmail(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	payload := `{"username": "trung", "email": "not-an-email"}`

	resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
	if err != nil {
		t.Fatalf("POST /users (invalid email) failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid email, got %d", resp.StatusCode)
	}
}

func TestGetAllUsers_WithDefaultData(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodGet, "/users", "", nil)
	if err != nil {
		t.Fatalf("GET /users failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}

	users := parseUsers(t, tr.Data)
	if len(users) != 1 {
		t.Fatalf("expected 1 default user, got %d", len(users))
	}
	if users[0].Username != "admin" {
		t.Fatalf("expected default username 'admin', got '%s'", users[0].Username)
	}
}

func TestGetAllUsers_AfterCreatingMultiple(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	usersToCreate := []string{
		`{"username": "user1", "email": "user1@test.com"}`,
		`{"username": "user2", "email": "user2@test.com"}`,
		`{"username": "user3", "email": "user3@test.com"}`,
	}

	for i, payload := range usersToCreate {
		resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
		if err != nil {
			t.Fatalf("creating user %d failed: %v", i, err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 for user %d, got %d", i, resp.StatusCode)
		}
		resp.Body.Close()
	}

	resp, err := doRequest(app, http.MethodGet, "/users", "", nil)
	if err != nil {
		t.Fatalf("GET /users failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}

	usersList := parseUsers(t, tr.Data)
	expectedTotal := 1 + len(usersToCreate)
	if len(usersList) != expectedTotal {
		t.Fatalf("expected %d users, got %d", expectedTotal, len(usersList))
	}

	usernames := make(map[string]bool)
	for _, u := range usersList {
		usernames[u.Username] = true
	}
	for _, expected := range []string{"user1", "user2", "user3"} {
		if !usernames[expected] {
			t.Fatalf("expected user '%s' in listing, not found", expected)
		}
	}
}

func TestGetUser_ByID(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	createPayload := `{"username": "getme", "email": "getme@test.com", "full_name": "Get Me"}`
	createResp, err := doRequest(app, http.MethodPost, "/users", createPayload, nil)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	tr := parseResponse(t, createResp)
	created := parseUser(t, tr.Data)

	getResp, err := doRequest(app, http.MethodGet, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("GET /users/%d failed: %v", created.ID, err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for GET /users/%d, got %d", created.ID, getResp.StatusCode)
	}

	tr = parseResponse(t, getResp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}

	user := parseUser(t, tr.Data)
	if user.ID != created.ID {
		t.Fatalf("expected user ID %d, got %d", created.ID, user.ID)
	}
	if user.Username != "getme" {
		t.Fatalf("expected username 'getme', got '%s'", user.Username)
	}
	if user.FullName != "Get Me" {
		t.Fatalf("expected full_name 'Get Me', got '%s'", user.FullName)
	}
}

func TestGetUser_DefaultUser(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodGet, "/users/1", "", nil)
	if err != nil {
		t.Fatalf("GET /users/1 failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for default user, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}

	user := parseUser(t, tr.Data)
	if user.Username != "admin" {
		t.Fatalf("expected username 'admin', got '%s'", user.Username)
	}
	if user.Email != "admin@example.com" {
		t.Fatalf("expected email 'admin@example.com', got '%s'", user.Email)
	}
}

func TestGetUser_NonExistent(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodGet, "/users/99999", "", nil)
	if err != nil {
		t.Fatalf("GET /users/99999 failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 400 or 404 for non-existent user, got %d", resp.StatusCode)
	}
}

func TestUpdateUser(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	createPayload := `{"username": "update_test", "email": "old@test.com", "full_name": "Before Update"}`
	createResp, err := doRequest(app, http.MethodPost, "/users", createPayload, nil)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	tr := parseResponse(t, createResp)
	created := parseUser(t, tr.Data)

	updatePayload := `{"email": "updated@test.com"}`
	updateResp, err := doRequest(app, http.MethodPatch, "/users/"+itoa(created.ID), updatePayload, nil)
	if err != nil {
		t.Fatalf("PATCH /users/%d failed: %v", created.ID, err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for update, got %d", updateResp.StatusCode)
	}

	tr = parseResponse(t, updateResp)
	if !tr.Success {
		t.Fatalf("expected success=true, got %+v", tr.Error)
	}

	getResp, err := doRequest(app, http.MethodGet, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("GET /users/%d after update failed: %v", created.ID, err)
	}

	tr = parseResponse(t, getResp)
	updated := parseUser(t, tr.Data)
	t.Logf("Updated user: %+v", updated)
	if updated.Username == created.Username {
		t.Log("Username preserved correctly")
	}
}

func TestUpdateUser_NonExistent(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	payload := `{"email": "nobody@test.com"}`
	resp, err := doRequest(app, http.MethodPatch, "/users/99999", payload, nil)
	if err != nil {
		t.Fatalf("PATCH /users/99999 failed: %v", err)
	}
	t.Logf("PATCH non-existent user returned status %d", resp.StatusCode)
}

func TestDeleteUser(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	createPayload := `{"username": "delete_me", "email": "delete@test.com"}`
	createResp, err := doRequest(app, http.MethodPost, "/users", createPayload, nil)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	tr := parseResponse(t, createResp)
	created := parseUser(t, tr.Data)

	deleteResp, err := doRequest(app, http.MethodDelete, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("DELETE /users/%d failed: %v", created.ID, err)
	}
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204 for delete, got %d", deleteResp.StatusCode)
	}

	getResp, err := doRequest(app, http.MethodGet, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("GET /users/%d after delete failed: %v", created.ID, err)
	}
	if getResp.StatusCode == http.StatusOK {
		t.Fatal("expected user to be gone after delete, but got 200")
	}
	t.Logf("GET deleted user returned status %d (expected error)", getResp.StatusCode)
}

func TestDeleteUser_NonExistent(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodDelete, "/users/99999", "", nil)
	if err != nil {
		t.Fatalf("DELETE /users/99999 failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204 for delete (idempotent), got %d", resp.StatusCode)
	}
}

func TestFullCRUDCycle(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	resp, err := doRequest(app, http.MethodGet, "/users", "", nil)
	if err != nil {
		t.Fatalf("initial GET /users failed: %v", err)
	}
	tr := parseResponse(t, resp)
	users := parseUsers(t, tr.Data)
	if len(users) < 1 {
		t.Fatalf("expected at least 1 default user, got %d", len(users))
	}

	createPayload := `{"username": "crud_user", "email": "crud@test.com", "full_name": "CRUD Tester"}`
	resp, err = doRequest(app, http.MethodPost, "/users", createPayload, nil)
	if err != nil {
		t.Fatalf("POST /users failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d", resp.StatusCode)
	}
	tr = parseResponse(t, resp)
	created := parseUser(t, tr.Data)
	if created.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if created.Username != "crud_user" {
		t.Fatalf("expected username 'crud_user', got '%s'", created.Username)
	}

	resp, err = doRequest(app, http.MethodGet, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("GET /users/%d failed: %v", created.ID, err)
	}
	tr = parseResponse(t, resp)
	fetched := parseUser(t, tr.Data)
	if fetched.Username != "crud_user" {
		t.Fatalf("expected username 'crud_user', got '%s'", fetched.Username)
	}

	updatePayload := `{"full_name": "CRUD Updated"}`
	resp, err = doRequest(app, http.MethodPatch, "/users/"+itoa(created.ID), updatePayload, nil)
	if err != nil {
		t.Fatalf("PATCH /users/%d failed: %v", created.ID, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}

	resp, err = doRequest(app, http.MethodDelete, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("DELETE /users/%d failed: %v", created.ID, err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d", resp.StatusCode)
	}

	resp, err = doRequest(app, http.MethodGet, "/users/"+itoa(created.ID), "", nil)
	if err != nil {
		t.Fatalf("post-delete GET /users/%d failed: %v", created.ID, err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Fatal("expected user to be gone after delete, but got 200")
	}
	t.Logf("GET deleted user returned status %d (expected error)", resp.StatusCode)
}

func TestCreateUser_WithContent(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	contentJSON := `{"nested": {"key": "value"}, "array": [1,2,3]}`
	payload := `{"username": "content_test", "email": "content@test.com", "content": ` + contentJSON + `}`

	resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
	if err != nil {
		t.Fatalf("POST /users with content failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	tr := parseResponse(t, resp)
	user := parseUser(t, tr.Data)

	if user.Content == nil {
		t.Fatal("expected non-nil content")
	}
}

func TestHealthCheck_MultipleRequests(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	for i := 0; i < 10; i++ {
		resp, err := doRequest(app, http.MethodGet, "/test", "", nil)
		if err != nil {
			t.Fatalf("iteration %d: GET /test failed: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("iteration %d: expected 200, got %d", i, resp.StatusCode)
		}

		tr := parseResponse(t, resp)
		if !tr.Success {
			t.Fatalf("iteration %d: expected success=true", i)
		}
		resp.Body.Close()
	}
}

func TestConcurrentCreateUsers(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	type result struct {
		index    int
		status   int
		username string
		err      error
	}

	results := make(chan result, 3)

	for i := 0; i < 3; i++ {
		go func(idx int) {
			username := "concurrent_user_" + itoa(int64(idx))
			payload := `{"username": "` + username + `", "email": "` + username + `@test.com"}`
			resp, err := doRequest(app, http.MethodPost, "/users", payload, nil)
			if err != nil {
				results <- result{index: idx, err: err}
				return
			}
			results <- result{index: idx, status: resp.StatusCode, username: username}
		}(i)
	}

	created := 0
	for i := 0; i < 3; i++ {
		r := <-results
		if r.err != nil {
			t.Fatalf("goroutine %d failed: %v", r.index, r.err)
		}
		if r.status == http.StatusCreated {
			created++
		} else {
			t.Logf("goroutine %d returned status %d", r.index, r.status)
		}
	}

	if created == 0 {
		t.Fatal("expected at least 1 concurrent create to succeed")
	}
	t.Logf("Successfully created %d/3 concurrent users", created)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
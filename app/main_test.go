/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type testRole struct {
	ID          uint64 `json:"id" otmc:"primary;auto"`
	Name        string `json:"name" otmc:"required;unique"`
	Description string `json:"description"`
}

type testUser struct {
	ID        uint64    `json:"id" otmc:"primary;auto"`
	Name      string    `json:"name" otmc:"required"`
	Email     string    `json:"email" otmc:"required;unique"`
	Password  string    `json:"-"` // json:"-" -> not part of the API schema
	RoleID    uint64    `json:"role_id"`
	Role      *testRole `json:"role,omitempty"`
	CreatedAt string    `json:"created_at"`
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	app := New(Config{
		Name:     "Test App",
		Database: DatabaseConfig{Driver: "sqlite", DSN: dsn},
	})
	app.Register(&testRole{}, &testUser{})
	if err := app.Initializer(); err != nil {
		t.Fatalf("Initializer: %v", err)
	}
	t.Cleanup(func() { app.Store().Close() })
	return app
}

func do(t *testing.T, app *App, method, path, body string) (*http.Response, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Fiber().Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

func TestCRUDLifecycle(t *testing.T) {
	app := newTestApp(t)

	// health
	resp, b := do(t, app, "GET", "/healthz", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"ok"`) {
		t.Fatalf("healthz: %d %s", resp.StatusCode, b)
	}

	// create role
	resp, b = do(t, app, "POST", "/api/test_role", `{"name":"admin","description":"Administrators"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create role: %d %s", resp.StatusCode, b)
	}
	if !strings.Contains(string(b), `"id":1`) {
		t.Fatalf("create role body: %s", b)
	}

	// unique violation -> 409
	resp, b = do(t, app, "POST", "/api/test_role", `{"name":"admin"}`)
	if resp.StatusCode != 409 {
		t.Fatalf("unique role: %d %s", resp.StatusCode, b)
	}

	// missing required -> 422
	resp, b = do(t, app, "POST", "/api/test_role", `{"description":"no name"}`)
	if resp.StatusCode != 422 {
		t.Fatalf("required role: %d %s", resp.StatusCode, b)
	}

	// create user
	resp, b = do(t, app, "POST", "/api/test_user", `{"name":"Alice","email":"a@x.com","role_id":1}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create user: %d %s", resp.StatusCode, b)
	}

	// duplicate email -> 409
	resp, b = do(t, app, "POST", "/api/test_user", `{"name":"Bob","email":"a@x.com"}`)
	if resp.StatusCode != 409 {
		t.Fatalf("unique email: %d %s", resp.StatusCode, b)
	}

	// list
	resp, b = do(t, app, "GET", "/api/test_user", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"total":1`) {
		t.Fatalf("list users: %d %s", resp.StatusCode, b)
	}
	if strings.Contains(string(b), "password") {
		t.Fatalf("hidden field leaked: %s", b)
	}

	// filter by field
	resp, b = do(t, app, "GET", "/api/test_user?email=a@x.com", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"total":1`) {
		t.Fatalf("filter users: %d %s", resp.StatusCode, b)
	}
	resp, b = do(t, app, "GET", "/api/test_user?email=none@x.com", "")
	if !strings.Contains(string(b), `"total":0`) {
		t.Fatalf("filter miss: %s", b)
	}

	// search
	resp, b = do(t, app, "GET", "/api/test_user/search?keyword=alice", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"total":1`) {
		t.Fatalf("search users: %d %s", resp.StatusCode, b)
	}

	// get
	resp, b = do(t, app, "GET", "/api/test_user/1", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"Alice"`) {
		t.Fatalf("get user: %d %s", resp.StatusCode, b)
	}

	// patch
	resp, b = do(t, app, "PATCH", "/api/test_user/1", `{"name":"Alicia"}`)
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"Alicia"`) {
		t.Fatalf("patch user: %d %s", resp.StatusCode, b)
	}

	// put (full update)
	resp, b = do(t, app, "PUT", "/api/test_user/1", `{"name":"Alice","email":"a@x.com","role_id":1}`)
	if resp.StatusCode != 200 {
		t.Fatalf("put user: %d %s", resp.StatusCode, b)
	}

	resp, b = do(t, app, "POST", "/api/test_user", `{"name":"Carol","email":"c@x.com"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create user2: %d %s", resp.StatusCode, b)
	}

	// update many
	resp, b = do(t, app, "PATCH", "/api/test_user?ids=1,2", `{"name":"Bulk"}`)
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"updated":2`) {
		t.Fatalf("update many: %d %s", resp.StatusCode, b)
	}
	resp, b = do(t, app, "GET", "/api/test_user/2", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"Bulk"`) {
		t.Fatalf("get patched user2: %d %s", resp.StatusCode, b)
	}

	// missing ids -> 400
	resp, b = do(t, app, "PATCH", "/api/test_user", `{"name":"X"}`)
	if resp.StatusCode != 400 {
		t.Fatalf("update many without ids: %d %s", resp.StatusCode, b)
	}

	// delete many
	resp, b = do(t, app, "DELETE", "/api/test_user?ids=2", "")
	if resp.StatusCode != 200 || !strings.Contains(string(b), `"deleted":1`) {
		t.Fatalf("delete many: %d %s", resp.StatusCode, b)
	}

	// delete
	resp, b = do(t, app, "DELETE", "/api/test_user/1", "")
	if resp.StatusCode != 200 {
		t.Fatalf("delete user: %d %s", resp.StatusCode, b)
	}
	resp, b = do(t, app, "GET", "/api/test_user/1", "")
	if resp.StatusCode != 404 {
		t.Fatalf("get deleted: %d %s", resp.StatusCode, b)
	}
}

type testAccount struct {
	ID       uint64 `json:"id" otmc:"primary;auto"`
	Username string `json:"username" otmc:"required;unique"`
	Password string `json:"password,omitempty" otmc:"hidden;hash=bcrypt"`
}

func storedPassword(t *testing.T, app *App, id uint64) string {
	t.Helper()
	var row struct {
		Password string
	}
	if err := app.Store().GormDB().Table("test_account").
		Where("id = ?", id).Take(&row).Error; err != nil {
		t.Fatalf("read stored password: %v", err)
	}
	return row.Password
}

func TestPasswordHashing(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "hash.db")
	app := New(Config{
		Name:     "Hash App",
		Database: DatabaseConfig{Driver: "sqlite", DSN: dsn},
	})
	app.Register(&testAccount{})
	if err := app.Initializer(); err != nil {
		t.Fatalf("Initializer: %v", err)
	}
	t.Cleanup(func() { app.Store().Close() })

	// create -> plaintext is bcrypt-hashed before insert
	resp, b := do(t, app, "POST", "/api/test_account", `{"username":"u1","password":"secret"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create account: %d %s", resp.StatusCode, b)
	}
	hash := storedPassword(t, app, 1)
	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("password not hashed: %q", hash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret")); err != nil {
		t.Fatalf("hash does not match plaintext: %v", err)
	}
	if strings.Contains(string(b), "secret") || strings.Contains(string(b), "$2") {
		t.Fatalf("password leaked in response: %s", b)
	}

	// patch without password -> stored hash untouched
	resp, b = do(t, app, "PATCH", "/api/test_account/1", `{"username":"u2"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("patch username: %d %s", resp.StatusCode, b)
	}
	if got := storedPassword(t, app, 1); got != hash {
		t.Fatalf("password changed on unrelated patch: %q -> %q", hash, got)
	}

	// patch with new plaintext -> re-hashed, old hash no longer matches
	resp, b = do(t, app, "PATCH", "/api/test_account/1", `{"password":"newsecret"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("patch password: %d %s", resp.StatusCode, b)
	}
	newHash := storedPassword(t, app, 1)
	if newHash == hash {
		t.Fatal("password was not re-hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte("newsecret")); err != nil {
		t.Fatalf("new hash does not match: %v", err)
	}

	// resending the stored hash is not double-hashed
	resp, b = do(t, app, "PATCH", "/api/test_account/1", `{"password":"`+newHash+`"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("patch hash: %d %s", resp.StatusCode, b)
	}
	if got := storedPassword(t, app, 1); got != newHash {
		t.Fatal("stored hash was re-hashed")
	}
}

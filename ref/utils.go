/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/otmc-sw/logger"

	db "otmc/app/db/sqlc"
	"otmc/app/llm"
	"otmc/app/scheduler"
	"otmc/app/settings"
)

type Handler struct {
	db             *db.Queries
	sqlDB          *sql.DB
	llmClient      *llm.Client
	settingsLoader *settings.Loader
	bkScheduler    *scheduler.BackupScheduler
}

func NewHandler(q *db.Queries, sqlDB *sql.DB, llmClient *llm.Client, settingsLoader *settings.Loader, bkScheduler *scheduler.BackupScheduler) *Handler {
	return &Handler{
		db:             q,
		sqlDB:          sqlDB,
		llmClient:      llmClient,
		settingsLoader: settingsLoader,
		bkScheduler:    bkScheduler,
	}
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error"`
}

type ErrorDetails struct {
	Code      int         `json:"code"`               // 400, 404, 500, ...
	Key       string      `json:"key"`                // BAD_REQUEST, NOT_FOUND, INTERNAL_SERVER_ERROR, ...
	Type      string      `json:"type"`               // Bad Request, Not Found, Backend Server Error
	Summary   string      `json:"summary"`            // User login failed
	Detail    string      `json:"detail"`             // Error message detail
	Reason    string      `json:"reason,omitempty"`   // Failure reason (if have)
	Request   interface{} `json:"request,omitempty"`  // Request body
	Data      interface{} `json:"data,omitempty"`     // Error data
	File      string      `json:"file,omitempty"`     // File name (e.g. main.go)
	Line      int         `json:"line,omitempty"`     // Line number
	Function  string      `json:"function,omitempty"` // Function name
	Timestamp string      `json:"timestamp"`          // Timestamp
}

func getCallerInfo() (file string, line int, funcName string) {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0, "unknown"
	}
	funcName = "unknown"
	if f := runtime.FuncForPC(pc); f != nil {
		parts := strings.Split(f.Name(), ".")
		funcName = parts[len(parts)-1]
	}
	return filepath.Base(file), line, funcName
}

func okJSON(c *fiber.Ctx, data interface{}) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func createdJSON(c *fiber.Ctx, data interface{}) error {
	c.Set("Content-Type", "application/json")
	c.Response().Header.Set("X-Content-Type-Options", "nosniff")
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func noContentJSON(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNoContent).JSON(SuccessResponse{
		Success: true,
		Data:    nil,
		Message: "Success",
	})
}

func mustJSON(data interface{}) []byte {
	var sb strings.Builder
	encoder := json.NewEncoder(&sb)
	encoder.SetEscapeHTML(false)
	encoder.Encode(data)
	return []byte(strings.TrimRight(sb.String(), "\n"))
}

func errJSON(c *fiber.Ctx, message string, err error) error {
	file, line, fn := getCallerInfo()
	timestamp := time.Now().Format(time.RFC3339)

	logger.Error("Request error: %s - %v", message, err)
	logger.Error("Location: %s:%d -> %s", file, line, fn)

	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:      fiber.StatusInternalServerError,
			Key:       "INTERNAL_ERROR",
			Type:      "Backend Server Error",
			Summary:   "Internal server error",
			Detail:    fmt.Sprintf("%s: %v", message, err),
			Reason:    err.Error(),
			Data:      nil,
			File:      file,
			Line:      line,
			Function:  fn,
			Timestamp: timestamp,
		},
	})
}

func badRequestJSON(c *fiber.Ctx, summary string, detail ...interface{}) error {
	file, line, fn := getCallerInfo()
	timestamp := time.Now().Format(time.RFC3339)

	logger.Warn("BAD REQUEST: %s", summary)
	logger.Warn("Location: %s:%d -> %s()", file, line, fn)
	if detail != nil {
		logger.Warn("Deatail: %+v", detail[0])
	}

	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:      fiber.StatusBadRequest,
			Key:       "BAD_REQUEST",
			Type:      "Bad Request",
			Summary:   summary,
			Detail:    fmt.Sprint(detail...),
			Reason:    summary,
			Request:   nil,
			Data:      nil,
			File:      file,
			Line:      line,
			Function:  fn,
			Timestamp: timestamp,
		},
	})
}

func notFoundJSON(c *fiber.Ctx, message string) error {
	file, line, fn := getCallerInfo()
	timestamp := time.Now().Format(time.RFC3339)

	logger.Warn("Resource not found: %s", message)
	logger.Warn("Location: %s:%d", file, line)

	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:      fiber.StatusNotFound,
			Key:       "NOT_FOUND",
			Type:      "Not Found",
			Summary:   "Resource not found",
			Detail:    message,
			Data:      nil,
			File:      file,
			Line:      line,
			Function:  fn,
			Timestamp: timestamp,
		},
	})
}

func unauthorizedJSON(c *fiber.Ctx, message string) error {
	file, line, fn := getCallerInfo()
	timestamp := time.Now().Format(time.RFC3339)

	logger.Warn("Unauthorized access: %s", message)
	logger.Warn("Location: %s:%d", file, line)

	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:      fiber.StatusUnauthorized,
			Key:       "UNAUTHORIZED",
			Type:      "Unauthorized",
			Summary:   "Unauthorized access",
			Detail:    message,
			Data:      nil,
			File:      file,
			Line:      line,
			Function:  fn,
			Timestamp: timestamp,
		},
	})
}

func validateAndExtractID(c *fiber.Ctx, param string) (string, error) {
	id := c.Params(param)
	if id == "" {
		return "", badRequestJSON(c, "ID parameter is required")
	}
	return id, nil
}

func validateAndExtractInt64ID(c *fiber.Ctx, param string) (int64, error) {
	idStr, err := validateAndExtractID(c, param)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, badRequestJSON(c, "ID must be a valid integer")
	}
	return id, nil
}

func stringOrNull(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func stringPtrOrNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func int64OrNull(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

func int64PtrOrNull(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

func float64OrNull(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{
		Float64: *f,
		Valid:   true,
	}
}

func timeToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

func nullTimeOrNull(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  t,
		Valid: true,
	}
}

func stringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func intToNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

func ptrInt64(i *int) int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return ""
}

func nullInt64ToInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

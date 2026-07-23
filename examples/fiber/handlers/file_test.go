/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	db "github.com/otmc-sw/rest/examples/fiber/db"
)

func setupFileTestApp(t *testing.T) (*fiber.App, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "rest-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	filesDir := filepath.Join(tmpDir, "data", "files")

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

	if err := os.MkdirAll(filesDir, 0755); err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create files directory: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		database.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to change directory: %v", err)
	}

	New(database.Queries)

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	app.Get("/download/file/:id", DownloadFile)
	app.Post("/upload/file", UploadFile)

	cleanup := func() {
		os.Chdir(originalDir)
		database.Close()
		os.RemoveAll(tmpDir)
	}

	return app, filesDir, cleanup
}

func createMultipartFormData(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestUploadFile_Success(t *testing.T) {
	app, filesDir, cleanup := setupFileTestApp(t)
	defer cleanup()

	fileContent := "This is test file content for upload"
	body, contentType := createMultipartFormData(t, "file", "test_upload.txt", fileContent)

	req, err := http.NewRequest(http.MethodPost, "/upload/file", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("POST /upload/file failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	testSubDir := filepath.Join(filesDir, "test")
	uploadedFiles, err := os.ReadDir(testSubDir)
	if err != nil {
		t.Fatalf("failed to read files directory: %v", err)
	}

	if len(uploadedFiles) == 0 {
		t.Fatal("expected at least 1 file to be uploaded")
	}

	savedPath := filepath.Join(testSubDir, uploadedFiles[0].Name())
	savedContent, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if string(savedContent) != fileContent {
		t.Fatalf("file content mismatch: expected '%s', got '%s'", fileContent, string(savedContent))
	}

	t.Logf("File uploaded successfully: %s", uploadedFiles[0].Name())
}

func TestUploadFile_NoFile(t *testing.T) {
	app, _, cleanup := setupFileTestApp(t)
	defer cleanup()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/upload/file", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("POST /upload/file failed: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Upload without file returned status %d", resp.StatusCode)
}

func TestDownloadFile_Success(t *testing.T) {
	app, filesDir, cleanup := setupFileTestApp(t)
	defer cleanup()

	testSubDir := filepath.Join(filesDir, "test")
	if err := os.MkdirAll(testSubDir, 0755); err != nil {
		t.Fatalf("failed to create test subdirectory: %v", err)
	}

	testFileName := "download_123.txt"
	testFilePath := filepath.Join(testSubDir, testFileName)
	testContent := "This is test file content for download"

	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/download/file/123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("GET /download/file failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	downloadedContent, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(downloadedContent) != testContent {
		t.Fatalf("downloaded content mismatch: expected '%s', got '%s'", testContent, string(downloadedContent))
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "attachment") {
		t.Logf("Content-Disposition: %s", contentDisposition)
	}

	t.Logf("File downloaded successfully")
}

func TestDownloadFile_NotFound(t *testing.T) {
	app, _, cleanup := setupFileTestApp(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodGet, "/download/file/999", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("GET /download/file failed: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Download non-existent file returned status %d", resp.StatusCode)
}

func TestUploadAndDownloadCycle(t *testing.T) {
	app, filesDir, cleanup := setupFileTestApp(t)
	defer cleanup()

	testSubDir := filepath.Join(filesDir, "test")
	if err := os.MkdirAll(testSubDir, 0755); err != nil {
		t.Fatalf("failed to create test subdirectory: %v", err)
	}

	uploadContent := "Test content for upload/download cycle"
	body, contentType := createMultipartFormData(t, "file", "cycle_test.txt", uploadContent)

	uploadReq, err := http.NewRequest(http.MethodPost, "/upload/file", body)
	if err != nil {
		t.Fatalf("failed to create upload request: %v", err)
	}
	uploadReq.Header.Set("Content-Type", contentType)

	uploadResp, err := app.Test(uploadReq, 5000)
	if err != nil {
		t.Fatalf("POST /upload/file failed: %v", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for upload, got %d", uploadResp.StatusCode)
	}

	testSubDir = filepath.Join(filesDir, "test")
	uploadedFiles, err := os.ReadDir(testSubDir)
	if err != nil {
		t.Fatalf("failed to read files directory: %v", err)
	}

	if len(uploadedFiles) == 0 {
		t.Fatal("expected file to be uploaded")
	}

	uploadedFileName := uploadedFiles[0].Name()
	t.Logf("Uploaded file: %s", uploadedFileName)

	uploadedPath := filepath.Join(testSubDir, uploadedFileName)
	downloadPath := filepath.Join(testSubDir, "download_123.txt")

	uploadedContent, err := os.ReadFile(uploadedPath)
	if err != nil {
		t.Fatalf("failed to read uploaded file: %v", err)
	}

	if err := os.WriteFile(downloadPath, uploadedContent, 0644); err != nil {
		t.Fatalf("failed to copy file to download location: %v", err)
	}

	downloadReq, err := http.NewRequest(http.MethodGet, "/download/file/123", nil)
	if err != nil {
		t.Fatalf("failed to create download request: %v", err)
	}

	downloadResp, err := app.Test(downloadReq, 5000)
	if err != nil {
		t.Fatalf("GET /download/file failed: %v", err)
	}
	defer downloadResp.Body.Close()

	downloadedContent, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		t.Fatalf("failed to read downloaded content: %v", err)
	}

	if string(downloadedContent) != uploadContent {
		t.Fatalf("content mismatch after cycle: expected '%s', got '%s'", uploadContent, string(downloadedContent))
	}

	t.Logf("Upload and download cycle completed successfully")
}

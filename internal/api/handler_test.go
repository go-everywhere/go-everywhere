package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jairo/assetter/internal/stability"
)

type mockStorage struct {
	files    map[string][]byte
	saveErr  error
	readErr  error
}

func (m *mockStorage) SaveFile(path string, data []byte) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	m.files[path] = data
	return nil
}

func (m *mockStorage) ReadFile(path string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	data, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func TestNewHandler(t *testing.T) {
	storage := &mockStorage{}
	// Create a real client with a test API key
	// In unit tests, we won't actually call the API
	client := stability.NewClient("test-api-key")
	
	handler := NewHandler(storage, client)
	
	if handler.storage == nil {
		t.Error("storage should not be nil")
	}
	if handler.client == nil {
		t.Error("client should not be nil")
	}
	if handler.jobs == nil {
		t.Error("jobs map should be initialized")
	}
}

func TestHandler_HomePage(t *testing.T) {
	handler := NewHandler(&mockStorage{}, stability.NewClient("test-key"))

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{
			name:       "GET request success",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST request not allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handler.HomePage(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "text/html" {
					t.Errorf("Expected Content-Type text/html, got %s", contentType)
				}
			}
		})
	}
}

func TestHandler_Upload(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		createFormFile func() (*bytes.Buffer, string)
		wantStatus     int
		wantErr        bool
	}{
		{
			name:   "Successful JPG upload",
			method: http.MethodPost,
			createFormFile: func() (*bytes.Buffer, string) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image data"))
				writer.Close()
				return &buf, writer.FormDataContentType()
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "Successful PNG upload",
			method: http.MethodPost,
			createFormFile: func() (*bytes.Buffer, string) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				part, _ := writer.CreateFormFile("image", "test.png")
				part.Write([]byte("fake png data"))
				writer.Close()
				return &buf, writer.FormDataContentType()
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "Invalid file type",
			method: http.MethodPost,
			createFormFile: func() (*bytes.Buffer, string) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				part, _ := writer.CreateFormFile("image", "test.txt")
				part.Write([]byte("not an image"))
				writer.Close()
				return &buf, writer.FormDataContentType()
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:   "GET request not allowed",
			method: http.MethodGet,
			createFormFile: func() (*bytes.Buffer, string) {
				return &bytes.Buffer{}, ""
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantErr:    true,
		},
		{
			name:   "Missing file",
			method: http.MethodPost,
			createFormFile: func() (*bytes.Buffer, string) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				writer.Close()
				return &buf, writer.FormDataContentType()
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&mockStorage{}, stability.NewClient("test-key"))
			
			body, contentType := tt.createFormFile()
			req := httptest.NewRequest(tt.method, "/upload", body)
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			
			w := httptest.NewRecorder()
			handler.Upload(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if !tt.wantErr {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				
				if response["job_id"] == "" {
					t.Error("Expected job_id in response")
				}
				if response["status"] != "processing" {
					t.Errorf("Expected status 'processing', got %s", response["status"])
				}
			}
		})
	}
}

func TestHandler_Status(t *testing.T) {
	handler := NewHandler(&mockStorage{}, stability.NewClient("test-key"))
	
	// Add a test job
	testJobID := "test-job-123"
	handler.jobs[testJobID] = &Job{
		ID:       testJobID,
		Status:   "processing",
		ModelURL: "",
	}

	tests := []struct {
		name       string
		method     string
		jobID      string
		wantStatus int
	}{
		{
			name:       "Get existing job status",
			method:     http.MethodGet,
			jobID:      testJobID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Get non-existent job status",
			method:     http.MethodGet,
			jobID:      "non-existent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "POST not allowed",
			method:     http.MethodPost,
			jobID:      testJobID,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/status/"+tt.jobID, nil)
			w := httptest.NewRecorder()

			handler.Status(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var job Job
				if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if job.ID != testJobID {
					t.Errorf("Expected job ID %s, got %s", testJobID, job.ID)
				}
			}
		})
	}
}

func TestHandler_Download(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "handler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create uploads directory
	uploadsDir := filepath.Join(tempDir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("Failed to create uploads dir: %v", err)
	}

	// Change to temp directory for the test
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	storage := &mockStorage{
		files: map[string][]byte{
			"completed-job.glb": []byte("GLB model data"),
		},
	}
	handler := NewHandler(storage, stability.NewClient("test-key"))
	
	// Create the actual file for http.ServeFile
	completedJobID := "completed-job"
	testFile := filepath.Join(uploadsDir, completedJobID+".glb")
	if err := os.WriteFile(testFile, []byte("GLB model data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Add test jobs
	handler.jobs[completedJobID] = &Job{
		ID:       completedJobID,
		Status:   "completed",
		ModelURL: "/download/" + completedJobID,
	}
	
	processingJobID := "processing-job"
	handler.jobs[processingJobID] = &Job{
		ID:     processingJobID,
		Status: "processing",
	}

	tests := []struct {
		name       string
		method     string
		jobID      string
		wantStatus int
	}{
		{
			name:       "Download completed model",
			method:     http.MethodGet,
			jobID:      completedJobID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Download processing job",
			method:     http.MethodGet,
			jobID:      processingJobID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Download non-existent job",
			method:     http.MethodGet,
			jobID:      "non-existent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "POST not allowed",
			method:     http.MethodPost,
			jobID:      completedJobID,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/download/"+tt.jobID, nil)
			w := httptest.NewRecorder()

			handler.Download(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// Test process3DGeneration is omitted as it would require mocking the stability client
// which is not straightforward with the current design.
// This would be better tested with integration tests.

func TestGenerateJobID(t *testing.T) {
	id1 := generateJobID()
	time.Sleep(1 * time.Nanosecond) // Ensure different timestamps
	id2 := generateJobID()
	
	if id1 == id2 {
		t.Error("Expected unique job IDs")
	}
	
	if !strings.HasPrefix(id1, "job_") {
		t.Error("Job ID should start with 'job_'")
	}
}
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jairo/assetter/internal/api"
	"github.com/jairo/assetter/internal/stability"
	"github.com/jairo/assetter/internal/storage"
	"github.com/jairo/assetter/testdata"
)

// TestServer tests the complete server workflow
func TestServer_CompleteWorkflow(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("STABILITY_API_KEY")
	if apiKey == "" {
		t.Skip("STABILITY_API_KEY not set, skipping integration test")
	}

	// Create temporary directory for uploads
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up components
	localStorage := storage.NewLocalStorage(filepath.Join(tempDir, "uploads"))
	stabilityClient := stability.NewClient(apiKey)
	handler := api.NewHandler(localStorage, stabilityClient)

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomePage)
	mux.HandleFunc("/upload", handler.Upload)
	mux.HandleFunc("/status/", handler.Status)
	mux.HandleFunc("/download/", handler.Download)
	
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test 1: Home page
	t.Run("HomePage", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/")
		if err != nil {
			t.Fatalf("Failed to get home page: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			t.Errorf("Expected HTML content type, got %s", resp.Header.Get("Content-Type"))
		}
	})

	// Test 2: Upload image and track job
	t.Run("UploadAndProcess", func(t *testing.T) {
		// Create test image
		imageData := testdata.CreateTestJPG(t)

		// Upload image
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("image", "test.jpg")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		
		if _, err := part.Write(imageData); err != nil {
			t.Fatalf("Failed to write image data: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		resp, err := http.Post(
			server.URL+"/upload",
			writer.FormDataContentType(),
			&buf,
		)
		if err != nil {
			t.Fatalf("Failed to upload: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Upload failed with status %d: %s", resp.StatusCode, body)
		}

		// Parse response to get job ID
		var uploadResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			t.Fatalf("Failed to decode upload response: %v", err)
		}

		jobID := uploadResp["job_id"]
		if jobID == "" {
			t.Fatal("No job ID in upload response")
		}

		// Poll job status
		var jobStatus map[string]interface{}
		maxAttempts := 120 // 10 minutes with 5-second intervals
		
		for i := 0; i < maxAttempts; i++ {
			resp, err := http.Get(server.URL + "/status/" + jobID)
			if err != nil {
				t.Fatalf("Failed to get job status: %v", err)
			}
			
			if err := json.NewDecoder(resp.Body).Decode(&jobStatus); err != nil {
				resp.Body.Close()
				t.Fatalf("Failed to decode status response: %v", err)
			}
			resp.Body.Close()

			status := jobStatus["status"].(string)
			t.Logf("Job %s status: %s (attempt %d/%d)", jobID, status, i+1, maxAttempts)

			if status == "completed" {
				// Test download
				modelURL := jobStatus["model_url"].(string)
				if modelURL == "" {
					t.Fatal("No model URL in completed job")
				}

				resp, err := http.Get(server.URL + modelURL)
				if err != nil {
					t.Fatalf("Failed to download model: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Download failed with status %d", resp.StatusCode)
				}

				// Verify it's GLB data
				data, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read model data: %v", err)
				}

				if len(data) == 0 {
					t.Error("Downloaded model is empty")
				}

				// GLB files start with "glTF"
				if len(data) >= 4 && string(data[:4]) != "glTF" {
					t.Error("Downloaded file doesn't appear to be a GLB model")
				}

				return // Success!
			} else if status == "failed" {
				errorMsg := ""
				if jobStatus["error"] != nil {
					errorMsg = jobStatus["error"].(string)
				}
				t.Fatalf("Job failed: %s", errorMsg)
			}

			time.Sleep(5 * time.Second)
		}

		t.Fatal("Job timed out")
	})
}

// TestServer_ErrorHandling tests various error scenarios
func TestServer_ErrorHandling(t *testing.T) {
	// Create temporary directory for uploads
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up components with invalid API key
	localStorage := storage.NewLocalStorage(filepath.Join(tempDir, "uploads"))
	stabilityClient := stability.NewClient("invalid-api-key")
	handler := api.NewHandler(localStorage, stabilityClient)

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", handler.Upload)
	mux.HandleFunc("/status/", handler.Status)
	mux.HandleFunc("/download/", handler.Download)
	
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test invalid file type
	t.Run("InvalidFileType", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("image", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		
		if _, err := part.Write([]byte("not an image")); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		resp, err := http.Post(
			server.URL+"/upload",
			writer.FormDataContentType(),
			&buf,
		)
		if err != nil {
			t.Fatalf("Failed to upload: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	// Test non-existent job
	t.Run("NonExistentJob", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/status/non-existent-job")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	// Test download of non-completed job
	t.Run("DownloadIncompleteJob", func(t *testing.T) {
		// First upload a file to get a job ID
		imageData := testdata.CreateTestJPG(t)
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		part, err := writer.CreateFormFile("image", "test.jpg")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		
		if _, err := part.Write(imageData); err != nil {
			t.Fatalf("Failed to write image data: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("Failed to close writer: %v", err)
		}

		resp, err := http.Post(
			server.URL+"/upload",
			writer.FormDataContentType(),
			&buf,
		)
		if err != nil {
			t.Fatalf("Failed to upload: %v", err)
		}
		
		var uploadResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			resp.Body.Close()
			t.Fatalf("Failed to decode response: %v", err)
		}
		resp.Body.Close()

		jobID := uploadResp["job_id"]

		// Try to download immediately (should fail)
		resp, err = http.Get(server.URL + "/download/" + jobID)
		if err != nil {
			t.Fatalf("Failed to download: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for incomplete job, got %d", resp.StatusCode)
		}
	})
}

// TestServer_ConcurrentRequests tests handling of concurrent requests
func TestServer_ConcurrentRequests(t *testing.T) {
	// Create temporary directory for uploads
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up components
	localStorage := storage.NewLocalStorage(filepath.Join(tempDir, "uploads"))
	stabilityClient := stability.NewClient("test-api-key") // Won't actually call API
	handler := api.NewHandler(localStorage, stabilityClient)

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.HomePage)
	mux.HandleFunc("/upload", handler.Upload)
	mux.HandleFunc("/status/", handler.Status)
	
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test concurrent uploads
	t.Run("ConcurrentUploads", func(t *testing.T) {
		numRequests := 10
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				imageData := testdata.CreateTestJPG(t)
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				part, err := writer.CreateFormFile("image", fmt.Sprintf("test%d.jpg", index))
				if err != nil {
					results <- err
					return
				}
				
				if _, err := part.Write(imageData); err != nil {
					results <- err
					return
				}
				
				if err := writer.Close(); err != nil {
					results <- err
					return
				}

				resp, err := http.Post(
					server.URL+"/upload",
					writer.FormDataContentType(),
					&buf,
				)
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("upload %d failed with status %d", index, resp.StatusCode)
					return
				}

				var uploadResp map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
					results <- err
					return
				}

				if uploadResp["job_id"] == "" {
					results <- fmt.Errorf("upload %d: no job ID in response", index)
					return
				}

				results <- nil
			}(i)
		}

		// Wait for all uploads to complete
		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent upload failed: %v", err)
			}
		}
	})
}
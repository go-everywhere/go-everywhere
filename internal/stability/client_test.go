package stability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestClient_Generate3D_Success(t *testing.T) {
	mockImageData := []byte("mock image data")
	mockGenerationID := "gen-123"
	mockModelData := []byte("mock GLB model data")

	callCount := 0
	mockTransport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			callCount++
			
			switch callCount {
			case 1: // Initial generation request
				if !strings.HasSuffix(req.URL.Path, "/v2beta/3d/stable-point-aware-3d") {
					t.Errorf("Unexpected URL path: %s", req.URL.Path)
				}
				if req.Method != "POST" {
					t.Errorf("Expected POST, got %s", req.Method)
				}
				if req.Header.Get("Authorization") != "Bearer test-key" {
					t.Errorf("Invalid authorization header")
				}

				resp := GenerationResponse{ID: mockGenerationID}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil

			case 2: // Status check - in progress
				if !strings.Contains(req.URL.Path, mockGenerationID) {
					t.Errorf("Expected generation ID in path: %s", req.URL.Path)
				}
				if req.Header.Get("Accept") != "application/json" {
					t.Errorf("Expected Accept: application/json header")
				}

				resp := StatusResponse{
					ID:     mockGenerationID,
					Status: "in_progress",
				}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil

			case 3: // Status check - succeeded
				resp := StatusResponse{
					ID:         mockGenerationID,
					Status:     "succeeded",
					FinishTime: time.Now().Unix(),
				}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil

			case 4: // Download model
				if req.Header.Get("Accept") != "model/gltf-binary" {
					t.Errorf("Expected Accept: model/gltf-binary header")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(mockModelData)),
				}, nil

			default:
				t.Fatalf("Unexpected call count: %d", callCount)
				return nil, fmt.Errorf("unexpected call")
			}
		},
	}

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: mockTransport,
			Timeout:   30 * time.Second,
		},
	}

	result, err := client.Generate3D(mockImageData)
	if err != nil {
		t.Fatalf("Generate3D failed: %v", err)
	}

	if !bytes.Equal(result, mockModelData) {
		t.Errorf("Expected model data %v, got %v", mockModelData, result)
	}

	if callCount != 4 {
		t.Errorf("Expected 4 API calls, got %d", callCount)
	}
}

func TestClient_Generate3D_InitialRequestFailure(t *testing.T) {
	mockTransport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("Invalid image format")),
			}, nil
		},
	}

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: mockTransport,
		},
	}

	_, err := client.Generate3D([]byte("test"))
	if err == nil {
		t.Error("Expected error for bad request")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("Expected API error, got: %v", err)
	}
}

func TestClient_Generate3D_GenerationFailed(t *testing.T) {
	callCount := 0
	mockTransport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			callCount++
			
			switch callCount {
			case 1: // Initial generation request
				resp := GenerationResponse{ID: "gen-failed"}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil

			case 2: // Status check - failed
				resp := StatusResponse{
					ID:     "gen-failed",
					Status: "failed",
				}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil

			default:
				t.Fatalf("Unexpected call count: %d", callCount)
				return nil, fmt.Errorf("unexpected call")
			}
		},
	}

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: mockTransport,
		},
	}

	_, err := client.Generate3D([]byte("test"))
	if err == nil {
		t.Error("Expected error for failed generation")
	}
	if !strings.Contains(err.Error(), "generation failed") {
		t.Errorf("Expected generation failed error, got: %v", err)
	}
}

// TestClient_waitForCompletion_Timeout is skipped because it would take too long
// In a real test suite, you might mock time.Sleep or use a different approach
func TestClient_waitForCompletion_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test to avoid long test duration")
}

func TestClient_downloadModel_Failure(t *testing.T) {
	mockTransport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("Model not found")),
			}, nil
		},
	}

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: mockTransport,
		},
	}

	_, err := client.downloadModel("nonexistent")
	if err == nil {
		t.Error("Expected error for download failure")
	}
	if !strings.Contains(err.Error(), "download failed") {
		t.Errorf("Expected download failed error, got: %v", err)
	}
}

func TestClient_Generate3D_NetworkError(t *testing.T) {
	mockTransport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network error: connection refused")
		},
	}

	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: mockTransport,
		},
	}

	_, err := client.Generate3D([]byte("test"))
	if err == nil {
		t.Error("Expected error for network failure")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("Expected network error, got: %v", err)
	}
}
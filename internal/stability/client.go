package stability

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	// Stability AI 3D generation endpoint - returns glTF binary data directly
	baseURL = "https://api.stability.ai/v2beta/3d/stable-point-aware-3d"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// These structs are no longer needed since the API returns the model directly
// type GenerationResponse struct {
// 	ID string `json:"id"`
// }
//
// type StatusResponse struct {
// 	ID         string `json:"id"`
// 	Status     string `json:"status"`
// 	FinishTime int64  `json:"finish_time,omitempty"`
// }

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) Generate3D(imageData []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		return nil, err
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("API endpoint not found (404). The endpoint %s may not exist or may have been deprecated. Response: %s", baseURL, string(body))
		}
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// The API returns the 3D model directly as binary glTF data
	// No need to parse JSON or wait for completion
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Validate that we received glTF binary data
	if len(body) < 4 || string(body[:4]) != "glTF" {
		return nil, fmt.Errorf("unexpected response format: expected glTF binary data, got: %s", string(body[:min(100, len(body))]))
	}

	fmt.Printf("Successfully received 3D model: %d bytes of glTF binary data\n", len(body))
	return body, nil
}

// These functions are no longer needed since the API returns the model directly
// Keeping them commented for reference:

// func (c *Client) waitForCompletion(generationID string) ([]byte, error) { ... }
// func (c *Client) downloadModel(generationID string) ([]byte, error) { ... }

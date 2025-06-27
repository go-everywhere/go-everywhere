package testdata

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"testing"
)

// CreateTestJPG creates a simple test JPEG image
func CreateTestJPG(t *testing.T) []byte {
	t.Helper()
	
	// Create a simple 100x100 red image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		t.Fatalf("Failed to create test JPEG: %v", err)
	}
	
	return buf.Bytes()
}

// CreateTestPNG creates a simple test PNG image
func CreateTestPNG(t *testing.T) []byte {
	t.Helper()
	
	// Create a simple 100x100 blue image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}
	
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		t.Fatalf("Failed to create test PNG: %v", err)
	}
	
	return buf.Bytes()
}

// CreateMockGLB creates mock GLB data for testing
func CreateMockGLB() []byte {
	// GLB file header (simplified for testing)
	// Real GLB files start with "glTF" magic bytes
	return []byte("glTF\x02\x00\x00\x00mock GLB content for testing")
}

// CompareBytes compares two byte slices and reports differences
func CompareBytes(t *testing.T, got, want []byte) {
	t.Helper()
	
	if !bytes.Equal(got, want) {
		t.Errorf("Byte slices differ:\ngot:  %v\nwant: %v", got, want)
	}
}

// DrainReader reads all data from a reader and returns it as bytes
func DrainReader(t *testing.T, r io.Reader) []byte {
	t.Helper()
	
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}
	
	return data
}
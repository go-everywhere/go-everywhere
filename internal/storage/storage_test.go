package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorage_SaveFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewLocalStorage(tempDir)

	tests := []struct {
		name    string
		path    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Save simple file",
			path:    "test.txt",
			data:    []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "Save file in subdirectory",
			path:    "subdir/test.txt",
			data:    []byte("hello from subdir"),
			wantErr: false,
		},
		{
			name:    "Save empty file",
			path:    "empty.txt",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "Save binary data",
			path:    "binary.bin",
			data:    []byte{0x00, 0xFF, 0x42},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.SaveFile(tt.path, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				fullPath := filepath.Join(tempDir, tt.path)
				savedData, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("Failed to read saved file: %v", err)
					return
				}

				if string(savedData) != string(tt.data) {
					t.Errorf("Saved data mismatch. Got %v, want %v", savedData, tt.data)
				}
			}
		})
	}
}

func TestLocalStorage_ReadFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewLocalStorage(tempDir)

	testData := []byte("test content")
	testPath := "test.txt"
	fullPath := filepath.Join(tempDir, testPath)

	if err := os.WriteFile(fullPath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    []byte
		wantErr bool
	}{
		{
			name:    "Read existing file",
			path:    testPath,
			want:    testData,
			wantErr: false,
		},
		{
			name:    "Read non-existent file",
			path:    "nonexistent.txt",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storage.ReadFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("ReadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalStorage_SaveAndReadIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewLocalStorage(tempDir)

	testCases := []struct {
		name string
		path string
		data []byte
	}{
		{"Text file", "docs/readme.txt", []byte("This is a readme")},
		{"Binary file", "models/test.glb", []byte{0x67, 0x6C, 0x54, 0x46}},
		{"JSON file", "config.json", []byte(`{"key": "value"}`)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := storage.SaveFile(tc.path, tc.data); err != nil {
				t.Fatalf("SaveFile failed: %v", err)
			}

			readData, err := storage.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("ReadFile failed: %v", err)
			}

			if string(readData) != string(tc.data) {
				t.Errorf("Data mismatch. Got %v, want %v", readData, tc.data)
			}
		})
	}
}

func TestNewLocalStorage_CreatesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	newDir := filepath.Join(tempDir, "new_storage_dir")
	storage := NewLocalStorage(newDir)

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("NewLocalStorage should create the directory if it doesn't exist")
	}

	testFile := "test.txt"
	testData := []byte("test")
	if err := storage.SaveFile(testFile, testData); err != nil {
		t.Errorf("Failed to save file in new directory: %v", err)
	}
}
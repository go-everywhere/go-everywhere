# 3D Model Generator

A web service that converts images to 3D models using the Stability AI API.

## Features

- Drag-and-drop image upload interface
- Real-time status updates
- 3D model generation using Stability AI's stable-point-aware-3d API
- Download generated models in GLB format
- **NEW: Interactive 3D model viewer with Three.js**
- **NEW: Model history with persistent storage**
- **NEW: Browse and view previously generated models**

## Prerequisites

- Go 1.21+
- Stability AI API key

## Setup

1. Clone the repository:
```bash
git clone https://github.com/jairo/assetter.git
cd assetter
```

2. Install dependencies:
```bash
go mod download
```

3. Set your Stability AI API key:
```bash
export STABILITY_API_KEY="your-api-key-here"
```

4. Run the server:
```bash
go run cmd/server/main.go
```

5. Open http://localhost:8080 in your browser

## Environment Variables

- `STABILITY_API_KEY` (required): Your Stability AI API key
- `PORT` (optional): Server port (default: 8080)

## Usage

1. Visit the web interface
2. Upload an image (JPG or PNG)
3. Wait for the 3D model to be generated
4. The model will automatically appear in the 3D viewer
5. Use mouse to rotate the model, scroll to zoom
6. Download the GLB file or browse previous models in the sidebar

## Project Structure

```
├── cmd/server/         # Main application entry point
├── internal/
│   ├── api/           # HTTP handlers
│   ├── models/        # Model metadata management
│   ├── storage/       # File storage interface
│   └── stability/     # Stability AI client
├── templates/         # Gomponents HTML templates
├── test/              # Integration tests
│   └── integration/   # Full workflow tests
├── testdata/          # Test utilities and fixtures
├── uploads/           # Generated 3D models storage
├── data/              # Model metadata storage
└── static/            # Static assets
```

## Testing

The project includes comprehensive unit and integration tests.

### Running Tests

```bash
# Run unit tests
make test

# Run integration tests (requires STABILITY_API_KEY)
make test-integration

# Run tests with coverage report
make test-coverage

# Run tests with detailed per-package coverage
make test-coverage-detailed
```

### Test Structure

- **Unit Tests**: Located alongside source files (`*_test.go`)
  - `internal/api/handler_test.go`: HTTP handler tests
  - `internal/stability/client_test.go`: API client tests with mocked HTTP
  - `internal/storage/storage_test.go`: File storage tests

- **Integration Tests**: Located in `test/integration/`
  - Full workflow testing (upload → process → download)
  - Error handling scenarios
  - Concurrent request handling

- **Test Utilities**: Located in `testdata/`
  - Helper functions for creating test images
  - Mock data generators

### Development Commands

```bash
# Format code
make fmt

# Run linters
make lint

# Run all checks (fmt, vet, lint, test)
make check

# Clean build artifacts and test files
make clean

# Show all available commands
make help
```
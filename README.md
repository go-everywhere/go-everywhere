# Go PWA Template

A modern Progressive Web Application (PWA) template built entirely in Go, featuring a distributed database backend and a WebAssembly-powered frontend.

## Architecture Overview

This template provides a full-stack Go application with:
- **Frontend**: Pure Go PWA using go-app framework compiled to WebAssembly
- **Backend**: HTTP server with RESTful API endpoints
- **Database**: Distributed in-memory database using Olric
- **Build System**: Dual compilation for server and WASM client

## Technology Stack

- **[go-app](https://go-app.dev/)**: Progressive Web App framework for Go
- **[Olric](https://github.com/olric-data/olric)**: Distributed in-memory data structure store
- **WebAssembly**: Client-side Go code compiled to WASM
- **Go 1.24.1**: Latest Go version with enhanced WASM support

## Project Structure

```
assette/
├── go/                     # Go application root
│   ├── main.go            # Server-side entry point (!js build tag)
│   ├── main_js.go         # Client-side entry point (js build tag)
│   ├── database.go        # Olric database configuration
│   ├── api/               # REST API endpoints
│   │   └── profile.go     # Profile API handler
│   ├── views/             # PWA page components
│   │   ├── home.go        # Home page view
│   │   └── profile.go     # Profile page view
│   ├── widgets/           # Reusable UI components
│   │   └── header.go      # Navigation header widget
│   ├── web/               # Compiled WASM output
│   │   └── app.wasm       # WebAssembly binary
│   └── Makefile           # Build commands
└── python/                # Alternative Python implementation (reference)

```

## Key Features

### 1. Progressive Web App (PWA)
- Single Page Application with client-side routing
- Offline capability through service workers
- Installable on desktop and mobile devices
- Responsive design ready

### 2. Distributed Database
- **Olric** distributed cache and storage
- Three deployment modes:
  - `local`: Single-node development mode
  - `lan`: Local area network clustering
  - `wan`: Wide area network clustering
- Embedded client for direct database access
- Automatic node discovery and data replication

### 3. Dual Compilation Strategy
- **Server Binary**: Full Go backend with database and API
- **WASM Binary**: Client-side Go compiled to WebAssembly
- Build tags (`//go:build js` and `//go:build !js`) for conditional compilation
- Shared view components between server and client

### 4. Component-Based Architecture
- Reusable UI widgets (Header, etc.)
- Composable view components
- Type-safe component properties
- Server-side rendering capability

## Getting Started

### Prerequisites
- Go 1.24.1 or higher
- Make (for build automation)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd assette
```

2. Install dependencies:
```bash
cd go
go mod download
```

### Building the Application

Build both server and WASM client:
```bash
cd go
make build
```

This will:
- Compile the WASM client to `web/app.wasm`
- Build the server binary to `tmp/main`

### Running the Application

Start the server:
```bash
cd go
./tmp/main
```

The application will be available at `http://localhost:8000`

## Development Workflow

### Adding New Pages

1. Create a new view component in `go/views/`:
```go
package views

import (
    "github.com/maxence-charriere/go-app/v10/pkg/app"
    "github.com/olric-data/olric"
)

type MyPage struct {
    app.Compo
    DB *olric.EmbeddedClient
}

func (p *MyPage) Render() app.UI {
    return app.Section().Body(
        // Your UI components
    )
}
```

2. Register routes in both `main.go` and `main_js.go`:
```go
// main.go (server-side)
app.Route("/mypage", func() app.Composer { 
    return &views.MyPage{DB: client} 
})

// main_js.go (client-side)
app.Route("/mypage", func() app.Composer { 
    return &views.MyPage{} 
})
```

### Adding API Endpoints

Create new handlers in `go/api/`:
```go
func MyHandler(db *olric.EmbeddedClient) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handler logic
    }
}
```

Register in `main.go`:
```go
http.HandleFunc("/api/myendpoint", api.MyHandler(client))
```

### Working with the Database

The Olric embedded client provides distributed key-value storage:
```go
// Get a distributed map
dm, err := client.NewDMap("users")

// Set a value
err = dm.Put(ctx, "user:123", userData)

// Get a value
value, err := dm.Get(ctx, "user:123")
```

## Configuration

### Database Modes

Configure in `database.go`:
- **local**: Single-node mode for development
- **lan**: Automatic discovery on local network
- **wan**: Manual configuration for internet clustering

### Server Configuration

- Port: Default 8000 (modify in `main.go`)
- PWA settings in `app.Handler` configuration

## Deployment

### Production Build

1. Build optimized binaries:
```bash
GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o web/app.wasm
go build -ldflags="-s -w" -o ./main .
```

2. Deploy with appropriate database configuration:
```go
// For production clustering
c := config.New("wan")
// Configure peer discovery
```

### Docker Deployment

Create a Dockerfile:
```dockerfile
FROM golang:1.24.1 AS builder
WORKDIR /app
COPY go/ .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tmp/main .
COPY --from=builder /app/web ./web
EXPOSE 8000
CMD ["./main"]
```

## Best Practices

1. **Component Design**: Keep components small and focused
2. **Database Access**: Use embedded client for read-heavy operations
3. **Build Tags**: Properly separate server and client code
4. **Error Handling**: Implement proper error handling in API endpoints
5. **Testing**: Write tests for both server and client components

## License

[Specify your license here]

## Contributing

[Add contribution guidelines]
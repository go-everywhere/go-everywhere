# Go PWA Template

A modern Progressive Web Application (PWA) template built entirely in Go, featuring embedded etcd for distributed storage and a WebAssembly-powered frontend.

## Architecture Overview

This template provides a full-stack Go application with:
- **Frontend**: Pure Go PWA using go-app framework compiled to WebAssembly
- **Backend**: HTTP server with RESTful API endpoints
- **Database**: Embedded etcd for distributed key-value storage
- **Build System**: Dual compilation for server and WASM client

## Technology Stack

- **[go-app](https://go-app.dev/)**: Progressive Web App framework for Go
- **[etcd](https://etcd.io/)**: Distributed reliable key-value store (embedded)
- **WebAssembly**: Client-side Go code compiled to WASM
- **Go 1.24.1**: Latest Go version with enhanced WASM support

## Project Structure

```
go-everywhere/
├── main.go            # Server-side entry point (!js build tag)
├── main_js.go         # Client-side entry point (js build tag)
├── database.go        # Embedded etcd configuration
├── api/               # REST API endpoints
│   ├── users.go       # User CRUD operations
│   └── message.go     # Message API handler
├── db/                # Database client layer
│   ├── client.go      # etcd client wrapper
│   └── errors.go      # Custom error types
├── models/            # Data models
│   └── user.go        # User model
├── views/             # PWA page components
│   ├── home.go        # Home page view
│   └── profile.go     # Profile page view
├── widgets/           # Reusable UI components
│   └── header.go      # Navigation header widget
├── web/               # Compiled WASM output
│   └── app.wasm       # WebAssembly binary
├── CLUSTER.md         # Clustering deployment guide
├── Makefile           # Build commands
└── README.md          # Project documentation
```

## Key Features

### 1. Progressive Web App (PWA)
- Single Page Application with client-side routing
- Offline capability through service workers
- Installable on desktop and mobile devices
- Responsive design ready

### 2. Embedded etcd Database
- **Embedded etcd** server runs within the application
- No external database required
- Automatic clustering support for multiple instances
- Strong consistency guarantees
- Built-in leader election and failover
- Key-value storage with namespaces

### 3. Dual Compilation Strategy
- **Server Binary**: Full Go backend with embedded etcd and API
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
git clone https://github.com/go-everywhere/go-everywhere.git
cd go-everywhere
```

2. Install dependencies:
```bash
go mod download
```

### Building the Application

Build both server and WASM client:
```bash
make build
```

This will:
- Compile the WASM client to `web/app.wasm`
- Build the server binary

### Running the Application

Start the server:
```bash
go run .
```

Or after building:
```bash
./go-everywhere
```

The application will be available at `http://localhost:8000`

## Development Workflow

### Adding New Pages

1. Create a new view component in `views/`:
```go
package views

import (
    "github.com/maxence-charriere/go-app/v10/pkg/app"
)

type MyPage struct {
    app.Compo
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
    return &views.MyPage{}
})

// main_js.go (client-side)
app.Route("/mypage", func() app.Composer {
    return &views.MyPage{}
})
```

### Adding API Endpoints

Create new handlers in `api/`:
```go
func MyHandler(client *db.Client) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handler logic using etcd client
    }
}
```

Register in `main.go`:
```go
http.HandleFunc("/api/myendpoint", api.MyHandler(client))
```

### Working with the Database

The etcd client wrapper provides simple key-value operations:
```go
// Store data
err := client.Put(ctx, "namespace", "key", value)

// Retrieve data
data, err := client.Get(ctx, "namespace", "key")

// Delete data
deleted, err := client.Delete(ctx, "namespace", "key")

// Get all keys in namespace
allData, err := client.GetAll(ctx, "namespace")
```

## Configuration

### Embedded etcd Configuration

The embedded etcd server starts automatically with the application. Configuration is in `database.go`:
- Data directory: Temporary directory by default
- Client port: 2379
- Peer port: 2380 (for clustering)

### Environment Variables (for clustering)

When deploying multiple instances, configure via environment:
- `ETCD_NAME`: Unique node name
- `ETCD_DATA_DIR`: Data directory path
- `ETCD_CLIENT_URLS`: Client listening URLs
- `ETCD_PEER_URLS`: Peer communication URLs
- `ETCD_INITIAL_CLUSTER`: Initial cluster configuration

See [CLUSTER.md](CLUSTER.md) for detailed clustering instructions.

## API Documentation

### User Management API

- `GET /api/users` - List all users
- `POST /api/users` - Create a new user
- `GET /api/users/{id}` - Get a specific user
- `PUT /api/users/{id}` - Update a user
- `DELETE /api/users/{id}` - Delete a user

### Message API

- `GET /api/message` - Get a sample message

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## Deployment

### Single Instance

1. Build the application:
```bash
go build -ldflags="-s -w" -o go-everywhere .
```

2. Run:
```bash
./go-everywhere
```

### Clustered Deployment

For high availability, deploy multiple instances. See [CLUSTER.md](CLUSTER.md) for:
- Local development clusters
- Docker Compose setup
- Kubernetes StatefulSet
- Production best practices

### Docker Deployment

Build and run with Docker:
```bash
docker build -t go-everywhere .
docker run -p 8000:8000 go-everywhere
```

Example Dockerfile:
```dockerfile
FROM golang:1.24.1 AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/go-everywhere .
COPY --from=builder /app/web ./web
EXPOSE 8000
CMD ["./go-everywhere"]
```

## Performance Considerations

### etcd Performance
- Embedded etcd handles thousands of operations per second
- For best performance, use SSD storage for data directory
- Monitor memory usage as data grows
- Consider data expiration policies for cache-like usage

### WebAssembly Optimization
- Minimize WASM binary size with build flags
- Use `-ldflags="-s -w"` to strip debug information
- Consider lazy loading for large applications

## Monitoring

### Health Check Endpoint

Add a health check for monitoring:
```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check etcd connectivity
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### Metrics
- etcd metrics available at `http://localhost:2379/metrics`
- Key metrics to monitor:
  - Leader elections
  - Storage size
  - Operation latencies

## Best Practices

1. **Component Design**: Keep components small and focused
2. **Database Access**: Use context for timeout control
3. **Error Handling**: Always handle etcd errors gracefully
4. **Build Tags**: Properly separate server and client code
5. **Testing**: Write tests for both server and client components
6. **Clustering**: Use odd number of nodes (3, 5, 7) for quorum
7. **Backup**: Regular backups of etcd data for disaster recovery

## Troubleshooting

### Common Issues

1. **Port already in use**: Change ports in configuration
2. **etcd fails to start**: Check data directory permissions
3. **WASM not loading**: Ensure `web/app.wasm` is built
4. **Cluster split-brain**: Ensure odd number of nodes

### Debug Mode

Enable debug logging:
```go
cfg.LogLevel = "debug"  // in database.go
```

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Specify your license here]

## Resources

- [go-app Documentation](https://go-app.dev/)
- [etcd Documentation](https://etcd.io/docs/)
- [WebAssembly with Go](https://github.com/golang/go/wiki/WebAssembly)
- [Clustering Guide](CLUSTER.md)
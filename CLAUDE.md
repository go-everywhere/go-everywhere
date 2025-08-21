# CLAUDE.md - Go PWA Template Development Guide

This document provides instructions for AI assistants to help create and organize applications following this Go PWA template structure.

## Template Overview

This is a full-stack Go Progressive Web Application template with:
- Pure Go frontend compiled to WebAssembly
- Go backend with distributed database (Olric)
- Component-based architecture using go-app framework
- Dual compilation strategy (server + WASM client)

## Project Structure Guidelines

### Core Directory Layout
```
project-root/
├── go/                     # All Go code
│   ├── main.go            # Server entry (!js build tag)
│   ├── main_js.go         # Client entry (js build tag)  
│   ├── database.go        # Olric setup (!js build tag)
│   ├── api/               # REST endpoints (!js build tag)
│   ├── views/             # Page components
│   ├── widgets/           # Reusable UI components
│   ├── models/            # Data models (if needed)
│   ├── services/          # Business logic (if needed)
│   ├── web/               # WASM output directory
│   └── Makefile           # Build automation
```

## Creating New Applications

### Step 1: Initial Setup

1. **Create project structure**:
```bash
mkdir -p myapp/go/{api,views,widgets,web}
cd myapp/go
```

2. **Initialize Go module**:
```bash
go mod init myapp
go get github.com/maxence-charriere/go-app/v10
go get github.com/olric-data/olric
```

3. **Create Makefile**:
```makefile
build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build -o ./tmp/main .

run: build
	./tmp/main

clean:
	rm -f web/app.wasm tmp/main
```

### Step 2: Core Files Creation

1. **Create `main.go`** (server-side):
```go
//go:build !js

package main

import (
    "myapp/api"
    "myapp/views"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
    db, client := database()
    
    // Register routes
    app.Route("/", func() app.Composer { return &views.Home{DB: client} })
    
    // Register API endpoints
    // http.HandleFunc("/api/example", api.ExampleHandler(client))
    
    http.Handle("/", &app.Handler{
        Name:        "MyApp",
        Description: "My Go PWA Application",
    })
    
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        if err := http.ListenAndServe(":8000", nil); err != nil {
            log.Fatal(err)
        }
    }()
    
    <-signalChan
    shutdown(db)
}
```

2. **Create `main_js.go`** (client-side):
```go
//go:build js

package main

import (
    "myapp/views"
    "github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
    app.Route("/", func() app.Composer { return &views.Home{} })
    app.RunWhenOnBrowser()
}
```

3. **Create `database.go`**:
```go
//go:build !js

package main

import (
    "context"
    "log"
    "time"
    
    "github.com/olric-data/olric"
    "github.com/olric-data/olric/config"
)

func database() (*olric.Olric, *olric.EmbeddedClient) {
    c := config.New("local") // Use "lan" or "wan" for clustering
    
    ctx, cancel := context.WithCancel(context.Background())
    c.Started = func() {
        defer cancel()
        log.Println("[INFO] Olric is ready")
    }
    
    db, err := olric.New(c)
    if err != nil {
        log.Fatalf("Failed to create Olric: %v", err)
    }
    
    go func() {
        if err := db.Start(); err != nil {
            log.Fatalf("olric.Start failed: %v", err)
        }
    }()
    
    <-ctx.Done()
    return db, db.NewEmbeddedClient()
}

func shutdown(db *olric.Olric) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := db.Shutdown(ctx); err != nil {
        log.Printf("Failed to shutdown Olric: %v", err)
    }
}
```

## Component Development Patterns

### Creating Views (Pages)

**Pattern for new page components**:
```go
package views

import (
    "github.com/maxence-charriere/go-app/v10/pkg/app"
    "github.com/olric-data/olric"
)

type PageName struct {
    app.Compo
    DB *olric.EmbeddedClient // Only on server-side
    
    // Component state
    loading bool
    data    []Item
}

func (p *PageName) OnMount(ctx app.Context) {
    // Initialize component
}

func (p *PageName) Render() app.UI {
    return app.Section().Body(
        // UI components
    )
}
```

### Creating Widgets (Reusable Components)

**Pattern for widgets**:
```go
package widgets

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type WidgetName struct {
    app.Compo
    
    // Props
    Title string
    OnClick func()
}

func (w *WidgetName) Render() app.UI {
    return app.Div().Body(
        // Widget UI
    )
}
```

### Creating API Endpoints

**Pattern for API handlers**:
```go
//go:build !js

package api

import (
    "encoding/json"
    "net/http"
    "github.com/olric-data/olric"
)

func HandlerName(db *olric.EmbeddedClient) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        
        // Handler logic
        
        json.NewEncoder(w).Encode(response)
    }
}
```

## Build Tag Strategy

### Server-Only Code (`//go:build !js`)
- Database connections
- API handlers
- Server configuration
- File I/O operations
- System calls

### Client-Only Code (`//go:build js`)
- WASM-specific logic
- Browser API calls
- Client-side routing

### Shared Code (no build tags)
- View components
- Widget components
- Data models
- Utility functions

## Common Tasks

### Adding a New Page

1. Create view file in `views/`:
```go
// views/about.go
package views

type About struct {
    app.Compo
}

func (a *About) Render() app.UI {
    return app.Section().Body(
        app.H1().Text("About Page"),
    )
}
```

2. Register in `main.go`:
```go
app.Route("/about", func() app.Composer { 
    return &views.About{DB: client} 
})
```

3. Register in `main_js.go`:
```go
app.Route("/about", func() app.Composer { 
    return &views.About{} 
})
```

### Adding Database Operations

1. Create a service layer:
```go
// services/user_service.go
//go:build !js

package services

import (
    "context"
    "github.com/olric-data/olric"
)

type UserService struct {
    client *olric.EmbeddedClient
}

func (s *UserService) GetUser(ctx context.Context, id string) (User, error) {
    dm, err := s.client.NewDMap("users")
    if err != nil {
        return User{}, err
    }
    
    val, err := dm.Get(ctx, id)
    // Process and return
}
```

### Implementing Real-time Features

Use go-app's state management:
```go
func (c *Component) OnMount(ctx app.Context) {
    ctx.Handle("event-name", c.handleEvent)
}

func (c *Component) handleEvent(ctx app.Context, a app.Action) {
    c.data = a.Value.(DataType)
    c.Update()
}

// Emit events
ctx.Emit("event-name", data)
```

## Best Practices

### 1. Component Organization
- One component per file
- Group related components in subdirectories
- Keep components focused and single-purpose

### 2. State Management
- Use component state for UI state
- Use context for shared state
- Use Olric for persistent state

### 3. API Design
- RESTful endpoints under `/api/`
- Return JSON responses
- Implement proper error handling

### 4. Database Usage
- Use embedded client for reads
- Implement caching strategies
- Handle network partitions gracefully

### 5. Build Optimization
- Use `-ldflags="-s -w"` for production builds
- Minimize WASM size by avoiding large dependencies
- Lazy load components when possible

## Testing Strategy

### Unit Tests
```go
// views/home_test.go
package views

import (
    "testing"
    "github.com/maxence-charriere/go-app/v10/pkg/app"
)

func TestHomeRender(t *testing.T) {
    home := &Home{}
    ui := home.Render()
    // Assert UI structure
}
```

### API Tests
```go
// api/handler_test.go
package api

import (
    "net/http/httptest"
    "testing"
)

func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/endpoint", nil)
    w := httptest.NewRecorder()
    
    handler := HandlerName(mockDB)
    handler(w, req)
    
    // Assert response
}
```

## Deployment Checklist

1. **Production Build**:
   - Optimize WASM size
   - Enable compression
   - Set production database mode

2. **Configuration**:
   - Environment variables for secrets
   - Database clustering setup
   - CORS configuration if needed

3. **Monitoring**:
   - Add logging middleware
   - Implement health checks
   - Set up metrics collection

4. **Security**:
   - Implement authentication
   - Add rate limiting
   - Enable HTTPS

## Common Patterns

### Loading States
```go
func (c *Component) Render() app.UI {
    if c.loading {
        return app.Div().Text("Loading...")
    }
    return app.Div().Body(/* content */)
}
```

### Error Handling
```go
func (c *Component) loadData(ctx app.Context) {
    ctx.Async(func() {
        data, err := fetchData()
        if err != nil {
            ctx.Dispatch(func() {
                c.error = err.Error()
                c.Update()
            })
            return
        }
        ctx.Dispatch(func() {
            c.data = data
            c.Update()
        })
    })
}
```

### Forms
```go
func (c *Component) Render() app.UI {
    return app.Form().OnSubmit(c.handleSubmit).Body(
        app.Input().
            Type("text").
            Value(c.input).
            OnInput(c.handleInput),
        app.Button().
            Type("submit").
            Text("Submit"),
    )
}
```

## Troubleshooting

### Common Issues

1. **WASM not loading**: Check build tags and Makefile
2. **Database connection errors**: Verify Olric configuration
3. **Routing issues**: Ensure routes registered in both main files
4. **API calls failing**: Check CORS and endpoint registration

### Debug Tips

1. Use browser DevTools for WASM debugging
2. Add logging to API handlers
3. Check Olric cluster status
4. Verify build tag placement

## Commands Reference

```bash
# Build everything
make build

# Run development server
make run

# Clean build artifacts
make clean

# Run tests
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## Notes for AI Assistants

When helping users with this template:

1. **Always check build tags** - Ensure code is in the right file based on whether it's server or client code
2. **Follow the pattern** - Use existing code as reference for consistency
3. **Component lifecycle** - Understand OnMount, OnDismount, and Update methods
4. **Database context** - Remember Olric is distributed and handle accordingly
5. **Dual compilation** - Keep in mind code runs both server-side and in WASM
6. **Type safety** - Leverage Go's type system for robust components
7. **Error handling** - Always implement proper error handling
8. **Performance** - Consider WASM size and runtime performance

This template provides a solid foundation for building modern, performant web applications entirely in Go.
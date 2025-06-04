# Wind Demo

This guide shows how to test the Wind CLI tool with the included example web application and real Go projects.

## Prerequisites

- Go 1.23+ installed
- Terminal/command line access

## Quick Demo

### 1. Build Wind

```bash
go build -o wind .
```

### 2. Test Wind CLI

```bash
# Show help
./wind help

# Show version
./wind version
```

### 3. Demo with Example Application

```bash
# Navigate to the example directory
cd example

# Start Wind watcher
../wind init
```

This will:

1. Auto-detect the project structure (cmd/api/ or main.go)
2. Build the web application with the correct command
3. Start the server on http://localhost:8080
4. Begin watching for file changes

### 4. Test Auto-Reload

With Wind running:

1. **Open your browser** to http://localhost:8080
2. **Edit example/main.go** - try changing the welcome message in the `homeHandler` function
3. **Save the file** - watch Wind automatically rebuild and restart the server
4. **Refresh your browser** - see your changes immediately

### 5. Example Changes to Try

**Change the welcome message:**

```go
// In homeHandler function, change:
<h1><span class="emoji">🌪️</span> Wind Example Server</h1>

// To:
<h1><span class="emoji">🌪️</span> Wind Example Server - UPDATED!</h1>
```

**Add a new endpoint:**

```go
// Add to main() function:
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello from Wind! Time: %s", time.Now().Format("15:04:05"))
})
```

**Change the port:**

```go
// Change:
port := ":8080"

// To:
port := ":3000"
```

### 6. Stop Wind

Press `Ctrl+C` to stop Wind. It will:

1. Gracefully shutdown the running application
2. Clean up temporary files
3. Exit cleanly

## Expected Output

When you start Wind, you should see colored output like:

```
🌪️  Starting Wind watcher...
Info: Detected project structure: Simple layout (root main.go)
Info: Current directory: /path/to/wind/example
🔨 Building application...
✅ Build successful
🚀 Starting application...
Success: Application started (PID: 12345)
🌪️ Wind Example Server starting on http://localhost:8080
Try editing this file and watch Wind reload automatically!
Press Ctrl+C to stop...
```

When you edit a file:

```
Change: File changed: main.go
🔨 Building application...
Info: Stopping application (PID: 12345)...
✅ Build successful
🚀 Starting application...
Success: Application started (PID: 12346)
🌪️ Wind Example Server starting on http://localhost:8080
Try editing this file and watch Wind reload automatically!
```

## Troubleshooting

### Port already in use

If you get "address already in use" error:

- Change the port in `example/main.go`
- Or kill any process using port 8080: `lsof -ti:8080 | xargs kill`

### Permission denied

Make sure Wind binary is executable:

```bash
chmod +x wind
```

### Build errors

Make sure the example compiles:

```bash
cd example
go build .
```

## Simple Usage

Wind now defaults to watching when run without arguments:

```bash
# From the example directory
../wind              # Starts watching immediately
../wind init         # Same as above
```

## Testing with Real Go Projects

Wind automatically works with standard Go project layouts:

### Production Projects (cmd/api/)

```bash
# Your typical production Go API structure:
your-api/
├── cmd/api/main.go    # Wind auto-detects this!
├── internal/handlers/
├── configs/
└── go.mod

# Just run Wind from the project root:
cd your-api
wind                   # Detects cmd/api/main.go automatically
```

### Expected Output for cmd/api/ projects:

```
🌪️  Starting Wind watcher...
Info: Detected project structure: Standard layout (cmd/api/)
Info: Current directory: /path/to/your-api
🔨 Building application...
✅ Build successful
🚀 Starting application...
Success: Application started (PID: 12346)
```

## Next Steps

- Try Wind with your own Go web applications
- Experiment with different file types (HTML, CSS, JS)
- Test with cmd/api/, cmd/, or root main.go structures
- Check out the source code to understand how it works
- Consider contributing features or improvements

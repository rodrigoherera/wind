# Wind - Go Web Application Watcher

🌪️ A fast and simple CLI tool for watching and auto-reloading Go web applications during development.

Wind is inspired by [air](https://github.com/air-verse/air) but focused specifically on web applications with a simpler, streamlined approach.

## Features

- ⚡ **Fast file watching** using polling with the Go standard library
- 🔄 **Automatic rebuild and reload** on file changes
- 🎨 **Colored output** using ANSI escape codes
- 🗂️ **Smart directory exclusion** (vendor, .git, node_modules, etc.)
- 📁 **Multiple file type support** (.go, .html, .css, .js, .json, .yaml, .yml)
- 🔧 **Graceful process management** with proper cleanup
- 🚀 **Zero configuration** - works out of the box
- 💻 **Simple CLI interface**
- 🏗️ **Zero dependencies** - uses only Go standard library
- 🎯 **Smart project detection** - automatically detects Go project layouts

## Installation

### Via `go install` (Recommended)

```bash
go install github.com/your-username/wind@latest
```

### Build from source

```bash
git clone https://github.com/your-username/wind.git
cd wind
go build -o wind .
```

## Usage

### Quick Start

Navigate to your Go web application directory and run:

```bash
wind init
```

This will:

1. Watch for file changes in the current directory
2. Automatically rebuild your application when files change
3. Restart the application with the new binary
4. Display colored output showing the build and run status

### Command Line Usage

```bash
wind              # Start watching current directory (default)
wind init         # Start watching current directory
wind help         # Show help message
wind version      # Show version
```

## How It Works

1. **Project Detection**: Automatically detects your Go project structure (cmd/api/, cmd/, or root main.go)
2. **File Watching**: Wind monitors your project directory using polling to detect file changes
3. **Smart Filtering**: Only reacts to relevant file types (.go, .html, .css, .js, etc.)
4. **Debouncing**: Groups rapid file changes to avoid unnecessary rebuilds
5. **Build Process**: Uses the appropriate build command based on your project structure
6. **Process Management**: Gracefully stops the previous process and starts the new one
7. **Cleanup**: Handles interrupts and cleans up temporary files

## Configuration

Wind works with zero configuration but uses sensible defaults:

- **Auto-detected Build Commands**:
  - `cmd/api/main.go` → `go build -o ./tmp/main ./cmd/api`
  - `cmd/main.go` → `go build -o ./tmp/main ./cmd`
  - `main.go` → `go build -o ./tmp/main .`
- **Run Command**: `./tmp/main`
- **Excluded Directories**: `vendor`, `.git`, `node_modules`, `tmp`, `.idea`, `.vscode`
- **Watched Extensions**: `.go`, `.html`, `.css`, `.js`, `.json`, `.yaml`, `.yml`
- **Poll Interval**: 500ms (file system polling)
- **Debounce Delay**: 300ms

## Supported Project Structures

Wind automatically detects and works with common Go project layouts:

### Standard Layout (Recommended)

```
your-web-app/
├── cmd/
│   └── api/
│       └── main.go     # Main application entry point
├── internal/           # Private application code
├── pkg/               # Public library code
├── configs/           # Configuration files
├── go.mod
├── go.sum
└── tmp/              # Created automatically for builds
    └── main          # Compiled binary
```

### Simple Layout

```
your-web-app/
├── main.go           # Main application file
├── handlers/         # Your HTTP handlers
├── models/          # Data models
├── go.mod
├── go.sum
└── tmp/            # Created automatically for builds
    └── main        # Compiled binary
```

### Alternative Layout

```
your-web-app/
├── cmd/
│   └── main.go      # Main application entry point
├── internal/        # Application code
├── go.mod
└── tmp/            # Created automatically for builds
    └── main        # Compiled binary
```

## Example Project

Here's a simple example of a Go web application that works great with Wind:

```go
// main.go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from Wind! 🌪️")
    })

    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Run Wind in the project directory:

```bash
wind init
```

Now edit `main.go` and watch Wind automatically rebuild and restart your server!

## Comparison with Air

Wind is inspired by [air](https://github.com/air-verse/air) but with key differences:

| Feature         | Wind               | Air                  |
| --------------- | ------------------ | -------------------- |
| Target Audience | Web applications   | General Go apps      |
| Configuration   | Zero-config        | Highly configurable  |
| Setup           | Works immediately  | Requires config file |
| File Types      | Web-focused        | Customizable         |
| Dependencies    | Zero (stdlib only) | Multiple external    |
| Size            | Lightweight        | Full-featured        |

Choose Wind if you want:

- Quick setup for web applications
- Zero configuration
- Lightweight tool with no dependencies
- Simple, focused functionality
- Single binary deployment

Choose Air if you need:

- Complex configuration options
- Support for various project types
- Advanced features
- Customizable workflows

## Troubleshooting

### "Permission denied" when running the binary

Make sure the binary is executable:

```bash
chmod +x ./tmp/main
```

### Files not being watched

Check if your files are in excluded directories. Wind excludes `vendor`, `.git`, `node_modules`, `tmp`, `.idea`, and `.vscode` by default.

### Build errors

Make sure your Go code compiles successfully:

```bash
go build .
```

Fix any compilation errors before running Wind.

## Development & Testing

Wind includes a comprehensive test suite to ensure reliability and performance.

### Running Tests

Wind comes with a full test suite including unit tests, integration tests, and benchmarks:

```bash
# Run all tests (unit + integration)
./test.sh

# Run only unit tests (fast)
./test.sh --unit-only

# Run only integration tests
./test.sh --integration-only

# Run performance benchmarks
./test.sh --benchmarks

# Generate test coverage report
./test.sh --coverage

# Run everything (tests, benchmarks, coverage)
./test.sh --all

# Verbose output
./test.sh --verbose

# Show help
./test.sh --help
```

### Test Coverage

The test suite covers:

- **Unit Tests**: Project structure detection, file filtering, change detection
- **Integration Tests**: Real file operations, complete workflows, error handling
- **Benchmark Tests**: Performance testing with various file counts
- **Performance Metrics**: File scanning ~97µs, change detection ~93µs

### Development Setup

```bash
# Clone the repository
git clone https://github.com/your-username/wind.git
cd wind

# Run tests to ensure everything works
./test.sh

# Build the binary
go build -o wind .

# Test with the example app
cd example
../wind init
```

All test results are saved in `test-results/` (ignored by git) for detailed analysis.

## Contributing

Contributions are welcome! Please:

1. Run the full test suite: `./test.sh --all`
2. Ensure all tests pass
3. Add tests for new features
4. Submit a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [air](https://github.com/air-verse/air)
- Built using only the Go standard library
- Uses polling-based file watching and ANSI escape codes for colors

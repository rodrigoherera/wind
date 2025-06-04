package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

type WindConfig struct {
	BuildCmd      string
	RunCmd        string
	ExcludeDirs   []string
	IncludeExts   []string
	PollInterval  time.Duration
	DebounceDelay time.Duration
}

type WindApp struct {
	config     WindConfig
	process    *os.Process
	building   bool
	mutex      sync.Mutex
	fileStates map[string]time.Time
	stopChan   chan bool
}

func main() {
	asciiWind := `Wind - Go Web App Watcher
 _    _ _____ _   _ _____  
| |  | |_   _| \ | |  __ \ 
| |  | | | | |  \| | |  | |
| |/\| | | | | . \ | |  | |
\  /\  /_| |_| |\  | |__| |
 \/  \/ \___/\_| \_|_____/ 
`
	fmt.Print(Cyan + asciiWind + Reset)

	// Default to init if no arguments provided
	if len(os.Args) == 1 {
		runWatcher()
		return
	}

	handleArgs(os.Args[1:])
}

func handleArgs(args []string) {
	switch args[0] {
	case "init":
		runWatcher()
	case "help", "-h", "--help":
		showHelp()
	case "version", "-v", "--version":
		fmt.Println("Wind v1.1.0 - Enhanced with smart project detection")
	default:
		fmt.Printf(Red+"Error: "+Reset+"Unknown command: %s\n", args[0])
		showHelp()
	}
}

func showHelp() {
	fmt.Printf(Cyan + "Wind - Go Web Application Watcher" + Reset + "\n")
	fmt.Println()
	fmt.Printf(Yellow + "Usage:" + Reset + "\n")
	fmt.Println("  wind              # Start watching current directory")
	fmt.Println("  wind init         # Start watching current directory")
	fmt.Println("  wind help         # Show this help message")
	fmt.Println("  wind version      # Show version")
	fmt.Println()
	fmt.Printf(Yellow + "Features:" + Reset + "\n")
	fmt.Println("  • Automatic reload on Go file changes")
	fmt.Println("  • Excludes common directories (vendor, .git, etc.)")
	fmt.Println("  • Colored output for better visibility")
	fmt.Println("  • Graceful process management")
	fmt.Println("  • Zero dependencies - uses only Go standard library")
}

func runWatcher() {
	// Auto-detect project structure and configure build command
	buildCmd, buildTarget := detectProjectStructure()

	config := WindConfig{
		BuildCmd:      buildCmd,
		RunCmd:        "./tmp/main",
		ExcludeDirs:   []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
		IncludeExts:   []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
		PollInterval:  500 * time.Millisecond,
		DebounceDelay: 300 * time.Millisecond,
	}

	fmt.Printf(Cyan+"Info: "+Reset+"Detected project structure: %s\n", buildTarget)

	app := &WindApp{
		config:     config,
		fileStates: make(map[string]time.Time),
		stopChan:   make(chan bool),
	}

	fmt.Printf(Green + "🌪️  Starting Wind watcher..." + Reset + "\n")
	fmt.Printf(Cyan+"Info: "+Reset+"Current directory: %s\n", getCurrentDir())

	// Create tmp directory if it doesn't exist
	if err := os.MkdirAll("tmp", 0755); err != nil {
		log.Printf(Red+"Error: "+Reset+"Failed to create tmp directory: %v", err)
		return
	}

	// Initial scan of files
	app.scanFiles()

	// Initial build and run
	app.buildAndRun()

	// Setup signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	fmt.Printf(Yellow + "Press Ctrl+C to stop..." + Reset + "\n")

	// Start file watching in a goroutine
	go app.watchFiles()

	// Wait for interrupt signal
	<-c
	fmt.Printf("\n" + Yellow + "Shutting down..." + Reset + "\n")
	close(app.stopChan)
	app.cleanup()
}

func (app *WindApp) scanFiles() error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		for _, exclude := range app.config.ExcludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Store file modification times
		if !info.IsDir() && app.shouldWatch(path) {
			app.fileStates[path] = info.ModTime()
		}

		return nil
	})
}

func (app *WindApp) watchFiles() {
	debounce := time.NewTimer(app.config.DebounceDelay)
	debounce.Stop()

	ticker := time.NewTicker(app.config.PollInterval)
	defer ticker.Stop()

	var hasChanges bool

	for {
		select {
		case <-app.stopChan:
			return

		case <-ticker.C:
			changed := app.checkForChanges()
			if changed && !hasChanges {
				hasChanges = true
				debounce.Reset(app.config.DebounceDelay)
			}

		case <-debounce.C:
			if hasChanges {
				hasChanges = false
				app.buildAndRun()
			}
		}
	}
}

func (app *WindApp) checkForChanges() bool {
	changed := false

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		for _, exclude := range app.config.ExcludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if file should be watched
		if !info.IsDir() && app.shouldWatch(path) {
			modTime := info.ModTime()
			if lastMod, exists := app.fileStates[path]; !exists || modTime.After(lastMod) {
				if exists {
					fmt.Printf(Yellow+"Change: "+Reset+"File changed: %s\n", path)
					changed = true
				}
				app.fileStates[path] = modTime
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf(Red+"Error: "+Reset+"Failed to scan files: %v\n", err)
	}

	return changed
}

func (app *WindApp) shouldWatch(filename string) bool {
	ext := filepath.Ext(filename)
	for _, includeExt := range app.config.IncludeExts {
		if ext == includeExt {
			return true
		}
	}
	return false
}

func (app *WindApp) buildAndRun() {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if app.building {
		return
	}
	app.building = true

	// Stop current process
	app.stopProcess()

	fmt.Printf(Cyan + "🔨 Building application..." + Reset + "\n")

	// Build the application
	buildCmd := exec.Command("sh", "-c", app.config.BuildCmd)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Printf(Red+"Error: "+Reset+"Build failed: %v\n", err)
		app.building = false
		return
	}

	fmt.Printf(Green + "✅ Build successful" + Reset + "\n")

	// Run the application
	fmt.Printf(Cyan + "🚀 Starting application..." + Reset + "\n")

	runCmd := exec.Command("sh", "-c", app.config.RunCmd)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	if err := runCmd.Start(); err != nil {
		fmt.Printf(Red+"Error: "+Reset+"Failed to start application: %v\n", err)
		app.building = false
		return
	}

	app.process = runCmd.Process
	fmt.Printf(Green+"Success: "+Reset+"Application started (PID: %d)\n", app.process.Pid)

	app.building = false
}

func (app *WindApp) stopProcess() {
	if app.process != nil {
		fmt.Printf(Yellow+"Info: "+Reset+"Stopping application (PID: %d)...\n", app.process.Pid)

		// Try graceful shutdown first
		if err := app.process.Signal(syscall.SIGTERM); err != nil {
			// Force kill if graceful shutdown fails
			app.process.Kill()
		}

		app.process.Wait()
		app.process = nil
	}
}

func (app *WindApp) cleanup() {
	app.stopProcess()

	// Clean up tmp directory
	if _, err := os.Stat("tmp/main"); err == nil {
		os.Remove("tmp/main")
	}
}

func detectProjectStructure() (buildCmd, buildTarget string) {
	// Check for standard Go project layouts

	// Option 1: cmd/api/main.go (most common for web APIs)
	if _, err := os.Stat("cmd/api/main.go"); err == nil {
		return "go build -o ./tmp/main ./cmd/api", "Standard layout (cmd/api/)"
	}

	// Option 2: cmd/main.go
	if _, err := os.Stat("cmd/main.go"); err == nil {
		return "go build -o ./tmp/main ./cmd", "Standard layout (cmd/)"
	}

	// Option 3: main.go in root (simple projects)
	if _, err := os.Stat("main.go"); err == nil {
		return "go build -o ./tmp/main .", "Simple layout (root main.go)"
	}

	// Option 4: Look for any main.go in cmd subdirectories
	if entries, err := os.ReadDir("cmd"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				mainPath := filepath.Join("cmd", entry.Name(), "main.go")
				if _, err := os.Stat(mainPath); err == nil {
					buildPath := "./cmd/" + entry.Name()
					return fmt.Sprintf("go build -o ./tmp/main %s", buildPath),
						fmt.Sprintf("Standard layout (cmd/%s/)", entry.Name())
				}
			}
		}
	}

	// Fallback to current directory
	return "go build -o ./tmp/main .", "Fallback (current directory)"
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

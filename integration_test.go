package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntegrationFullWorkflow tests the complete Wind workflow
func TestIntegrationFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := createTempProject(t, "cmd-api")
	defer os.RemoveAll(tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create WindApp
	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
			ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
		},
		fileStates: make(map[string]time.Time),
	}

	// Test initial build
	buildCmd, _ := detectProjectStructure()
	if !strings.Contains(buildCmd, "./cmd/api") {
		t.Errorf("Expected build command to contain './cmd/api', got: %s", buildCmd)
	}

	// Test file scanning
	if err := app.scanFiles(); err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	// Should find the main.go file
	mainFile := filepath.Join("cmd", "api", "main.go")
	if _, exists := app.fileStates[mainFile]; !exists {
		t.Error("main.go should be tracked in fileStates")
	}

	// Test build process (basic compilation check)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parts := strings.Fields(buildCmd)
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = tmpDir

	// Create tmp directory first
	if err := os.MkdirAll("tmp", 0755); err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}

	if err := cmd.Run(); err != nil {
		t.Fatalf("Build command failed: %v", err)
	}

	// Check if binary was created
	binaryPath := "./tmp/main"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Error("Binary was not created after build")
	}
}

// TestIntegrationFileWatching tests file change detection
func TestIntegrationFileWatching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := createTempProject(t, "root")
	defer os.RemoveAll(tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
			ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
		},
		fileStates: make(map[string]time.Time),
	}

	// Initial scan
	if err := app.scanFiles(); err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	// Wait a bit to ensure different timestamps
	time.Sleep(100 * time.Millisecond)

	// Create a new Go file
	newFileContent := `package main

import "fmt"

func helper() {
	fmt.Println("Helper function")
}`

	if err := os.WriteFile("helper.go", []byte(newFileContent), 0644); err != nil {
		t.Fatalf("Failed to create helper.go: %v", err)
	}

	// Should detect the new file on next scan
	if changed := app.checkForChanges(); !changed {
		// New files are tracked but don't trigger change on first detection
		// This is the expected behavior - the file gets added to fileStates
		t.Log("New file detected and added to tracking (expected behavior)")
	}

	// Verify new file is tracked
	if _, exists := app.fileStates["helper.go"]; !exists {
		t.Error("New file should be tracked")
	}

	// Wait a bit more
	time.Sleep(100 * time.Millisecond)

	// Modify existing file
	modifiedContent := `package main

import "fmt"

func main() {
	fmt.Println("Modified main function")
	helper()
}

func helper() {
	fmt.Println("Helper function")
}`

	if err := os.WriteFile("main.go", []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify main.go: %v", err)
	}

	// Should detect the modification
	if changed := app.checkForChanges(); !changed {
		t.Error("Should detect file modification")
	}
}

// TestIntegrationMultipleProjectTypes tests different project structures
func TestIntegrationMultipleProjectTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectTypes := []struct {
		name      string
		structure string
		mainFile  string
	}{
		{"root_project", "root", "main.go"},
		{"cmd_project", "cmd-root", filepath.Join("cmd", "main.go")},
		{"api_project", "cmd-api", filepath.Join("cmd", "api", "main.go")},
		{"custom_project", "cmd-custom", filepath.Join("cmd", "server", "main.go")},
	}

	for _, pt := range projectTypes {
		t.Run(pt.name, func(t *testing.T) {
			tmpDir := createTempProject(t, pt.structure)
			defer os.RemoveAll(tmpDir)

			originalDir, _ := os.Getwd()
			defer func() { os.Chdir(originalDir) }()

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp dir: %v", err)
			}

			app := &WindApp{
				config: WindConfig{
					IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
					ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
				},
				fileStates: make(map[string]time.Time),
			}

			// Test file scanning
			if err := app.scanFiles(); err != nil {
				t.Fatalf("Failed to scan files for %s: %v", pt.name, err)
			}

			// Should find the main.go file
			if _, exists := app.fileStates[pt.mainFile]; !exists {
				t.Errorf("main.go should be tracked for %s at %s", pt.name, pt.mainFile)
			}

			// Test project structure detection
			buildCmd, buildTarget := detectProjectStructure()
			if buildCmd == "" {
				t.Errorf("Build command should not be empty for %s", pt.name)
			}
			if buildTarget == "" {
				t.Errorf("Build target should not be empty for %s", pt.name)
			}
		})
	}
}

// TestIntegrationErrorHandling tests error scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with invalid Go project (no go.mod)
	tmpDir, err := os.MkdirTemp("", "wind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create only main.go without go.mod
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// This should detect simple layout since main.go exists
	buildCmd, buildTarget := detectProjectStructure()
	if buildCmd != "go build -o ./tmp/main ." {
		t.Errorf("Expected build command, got: %s", buildCmd)
	}
	if buildTarget != "Simple layout (root main.go)" {
		t.Errorf("Expected simple layout target, got: %s", buildTarget)
	}
}

// TestIntegrationPerformance tests performance with many files
func TestIntegrationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := createTempProject(t, "root")
	defer os.RemoveAll(tmpDir)

	// Create many Go files
	for i := 0; i < 100; i++ {
		content := fmt.Sprintf(`package main

import "fmt"

func function%d() {
	fmt.Println("Function %d")
}`, i, i)

		filename := fmt.Sprintf("file%d.go", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
			ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
		},
		fileStates: make(map[string]time.Time),
	}

	// Measure scan time
	start := time.Now()
	if err := app.scanFiles(); err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}
	scanDuration := time.Since(start)

	// Should have scanned all 101 files (100 + main.go)
	expectedFiles := 101
	if len(app.fileStates) != expectedFiles {
		t.Errorf("Expected %d files, got %d", expectedFiles, len(app.fileStates))
	}

	// Scan should be reasonably fast (under 1 second for 100 files)
	if scanDuration > time.Second {
		t.Errorf("Scan took too long: %v", scanDuration)
	}

	// Measure change detection time
	start = time.Now()
	changed := app.checkForChanges()
	checkDuration := time.Since(start)

	// No changes expected
	if changed {
		t.Error("No changes should be detected")
	}

	// Check should be fast
	if checkDuration > 100*time.Millisecond {
		t.Errorf("Change check took too long: %v", checkDuration)
	}

	t.Logf("Performance: Scan %d files in %v, check changes in %v", expectedFiles, scanDuration, checkDuration)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test helper functions
func createTempProject(t *testing.T, structure string) string {
	tmpDir, err := os.MkdirTemp("", "wind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	switch structure {
	case "cmd-api":
		// Create cmd/api/main.go structure
		if err := os.MkdirAll(filepath.Join(tmpDir, "cmd", "api"), 0755); err != nil {
			t.Fatalf("Failed to create cmd/api dir: %v", err)
		}
		mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from cmd/api")
}`
		if err := os.WriteFile(filepath.Join(tmpDir, "cmd", "api", "main.go"), []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.go: %v", err)
		}

	case "cmd-root":
		// Create cmd/main.go structure
		if err := os.MkdirAll(filepath.Join(tmpDir, "cmd"), 0755); err != nil {
			t.Fatalf("Failed to create cmd dir: %v", err)
		}
		mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from cmd")
}`
		if err := os.WriteFile(filepath.Join(tmpDir, "cmd", "main.go"), []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.go: %v", err)
		}

	case "root":
		// Create root main.go structure
		mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from root")
}`
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.go: %v", err)
		}

	case "cmd-custom":
		// Create cmd/server/main.go structure
		if err := os.MkdirAll(filepath.Join(tmpDir, "cmd", "server"), 0755); err != nil {
			t.Fatalf("Failed to create cmd/server dir: %v", err)
		}
		mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from cmd/server")
}`
		if err := os.WriteFile(filepath.Join(tmpDir, "cmd", "server", "main.go"), []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.go: %v", err)
		}

	case "empty":
		// Empty project with no main.go
		// Just create the directory

	default:
		t.Fatalf("Unknown project structure: %s", structure)
	}

	// Create go.mod for all projects
	goModContent := `module test-project

go 1.23.4`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	return tmpDir
}

func TestDetectProjectStructure(t *testing.T) {
	tests := []struct {
		name           string
		structure      string
		expectedCmd    string
		expectedTarget string
	}{
		{
			name:           "cmd/api structure",
			structure:      "cmd-api",
			expectedCmd:    "go build -o ./tmp/main ./cmd/api",
			expectedTarget: "Standard layout (cmd/api/)",
		},
		{
			name:           "cmd root structure",
			structure:      "cmd-root",
			expectedCmd:    "go build -o ./tmp/main ./cmd",
			expectedTarget: "Standard layout (cmd/)",
		},
		{
			name:           "root main.go structure",
			structure:      "root",
			expectedCmd:    "go build -o ./tmp/main .",
			expectedTarget: "Simple layout (root main.go)",
		},
		{
			name:           "cmd/server custom structure",
			structure:      "cmd-custom",
			expectedCmd:    "go build -o ./tmp/main ./cmd/server",
			expectedTarget: "Standard layout (cmd/server/)",
		},
		{
			name:           "empty project fallback",
			structure:      "empty",
			expectedCmd:    "go build -o ./tmp/main .",
			expectedTarget: "Fallback (current directory)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTempProject(t, tt.structure)
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp dir: %v", err)
			}

			buildCmd, buildTarget := detectProjectStructure()

			if buildCmd != tt.expectedCmd {
				t.Errorf("Expected buildCmd %q, got %q", tt.expectedCmd, buildCmd)
			}

			if buildTarget != tt.expectedTarget {
				t.Errorf("Expected buildTarget %q, got %q", tt.expectedTarget, buildTarget)
			}
		})
	}
}

func TestShouldWatch(t *testing.T) {
	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
		},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"main.go", true},
		{"handler.go", true},
		{"template.html", true},
		{"style.css", true},
		{"script.js", true},
		{"config.json", true},
		{"config.yaml", true},
		{"config.yml", true},
		{"README.md", false},
		{"image.png", false},
		{"data.xml", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := app.shouldWatch(tt.filename)
			if result != tt.expected {
				t.Errorf("shouldWatch(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFileStateTracking(t *testing.T) {
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

	// Should have main.go in file states
	if _, exists := app.fileStates["main.go"]; !exists {
		t.Error("main.go should be tracked in fileStates")
	}

	// Modify main.go
	time.Sleep(time.Millisecond * 10) // Ensure different timestamp
	newContent := `package main

import "fmt"

func main() {
	fmt.Println("Modified content")
}`
	if err := os.WriteFile("main.go", []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to modify main.go: %v", err)
	}

	// Check for changes should detect the modification
	changed := app.checkForChanges()
	if !changed {
		t.Error("checkForChanges should detect the modification")
	}
}

func TestExcludeDirectories(t *testing.T) {
	tmpDir := createTempProject(t, "root")
	defer os.RemoveAll(tmpDir)

	// Create excluded directories with Go files
	excludedDirs := []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"}
	for _, dir := range excludedDirs {
		dirPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}

		goFile := filepath.Join(dirPath, "test.go")
		if err := os.WriteFile(goFile, []byte("package main"), 0644); err != nil {
			t.Fatalf("Failed to write file in %s: %v", dir, err)
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
			ExcludeDirs: excludedDirs,
		},
		fileStates: make(map[string]time.Time),
	}

	if err := app.scanFiles(); err != nil {
		t.Fatalf("Failed to scan files: %v", err)
	}

	// Should only track main.go, not files in excluded directories
	expectedFiles := 1 // Only main.go
	if len(app.fileStates) != expectedFiles {
		t.Errorf("Expected %d tracked files, got %d", expectedFiles, len(app.fileStates))
	}

	// Verify main.go is tracked
	if _, exists := app.fileStates["main.go"]; !exists {
		t.Error("main.go should be tracked")
	}

	// Verify excluded files are not tracked
	for _, dir := range excludedDirs {
		excludedFile := filepath.Join(dir, "test.go")
		if _, exists := app.fileStates[excludedFile]; exists {
			t.Errorf("File %s should not be tracked (excluded directory)", excludedFile)
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	tmpDir := createTempProject(t, "cmd-api")
	defer os.RemoveAll(tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	buildCmd, buildTarget := detectProjectStructure()

	expectedBuildCmd := "go build -o ./tmp/main ./cmd/api"
	expectedBuildTarget := "Standard layout (cmd/api/)"

	if buildCmd != expectedBuildCmd {
		t.Errorf("Expected buildCmd %q, got %q", expectedBuildCmd, buildCmd)
	}

	if buildTarget != expectedBuildTarget {
		t.Errorf("Expected buildTarget %q, got %q", expectedBuildTarget, buildTarget)
	}
}

func TestMultipleMainFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wind-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple main.go files in different cmd subdirectories
	subDirs := []string{"api", "worker", "migrator"}
	for _, subDir := range subDirs {
		dirPath := filepath.Join(tmpDir, "cmd", subDir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dirPath, err)
		}

		mainContent := fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("Hello from %s")
}`, subDir)
		mainFile := filepath.Join(dirPath, "main.go")
		if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
			t.Fatalf("Failed to write main.go in %s: %v", subDir, err)
		}
	}

	// Create go.mod
	goModContent := `module test-project

go 1.23.4`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	buildCmd, buildTarget := detectProjectStructure()

	// Should detect the first one found (alphabetically, "api" comes first)
	expectedBuildCmd := "go build -o ./tmp/main ./cmd/api"
	expectedBuildTarget := "Standard layout (cmd/api/)"

	if buildCmd != expectedBuildCmd {
		t.Errorf("Expected buildCmd %q, got %q", expectedBuildCmd, buildCmd)
	}

	if buildTarget != expectedBuildTarget {
		t.Errorf("Expected buildTarget %q, got %q", expectedBuildTarget, buildTarget)
	}
}

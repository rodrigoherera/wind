package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkScanFiles benchmarks file scanning performance
func BenchmarkScanFiles(b *testing.B) {
	tmpDir := createTempProject(&testing.T{}, "root")
	defer os.RemoveAll(tmpDir)

	// Create additional files for benchmarking
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf(`package main

import "fmt"

func function%d() {
	fmt.Println("Function %d")
}`, i, i)

		filename := fmt.Sprintf("file%d.go", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp dir: %v", err)
	}

	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
			ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
		},
		fileStates: make(map[string]time.Time),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.fileStates = make(map[string]time.Time) // Reset state
		if err := app.scanFiles(); err != nil {
			b.Fatalf("Failed to scan files: %v", err)
		}
	}
}

// BenchmarkCheckForChanges benchmarks change detection performance
func BenchmarkCheckForChanges(b *testing.B) {
	tmpDir := createTempProject(&testing.T{}, "root")
	defer os.RemoveAll(tmpDir)

	// Create additional files for benchmarking
	for i := 0; i < 50; i++ {
		content := fmt.Sprintf(`package main

import "fmt"

func function%d() {
	fmt.Println("Function %d")
}`, i, i)

		filename := fmt.Sprintf("file%d.go", i)
		if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp dir: %v", err)
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
		b.Fatalf("Failed to scan files: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.checkForChanges()
	}
}

// BenchmarkDetectProjectStructure benchmarks project structure detection
func BenchmarkDetectProjectStructure(b *testing.B) {
	structures := []string{"root", "cmd-root", "cmd-api", "cmd-custom"}

	for _, structure := range structures {
		b.Run(structure, func(b *testing.B) {
			tmpDir := createTempProject(&testing.T{}, structure)
			defer os.RemoveAll(tmpDir)

			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				b.Fatalf("Failed to change to temp dir: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				detectProjectStructure()
			}
		})
	}
}

// BenchmarkShouldWatch benchmarks file extension checking
func BenchmarkShouldWatch(b *testing.B) {
	app := &WindApp{
		config: WindConfig{
			IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
		},
	}

	testFiles := []string{
		"main.go", "handler.go", "template.html", "style.css", "script.js",
		"config.json", "config.yaml", "config.yml", "README.md", "image.png",
		"data.xml", "test.txt", "archive.zip", "binary.exe",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, file := range testFiles {
			app.shouldWatch(file)
		}
	}
}

// BenchmarkFileScanning benchmarks scanning with different file counts
func BenchmarkFileScanning(b *testing.B) {
	fileCounts := []int{10, 50, 100, 500}

	for _, count := range fileCounts {
		b.Run(fmt.Sprintf("files_%d", count), func(b *testing.B) {
			tmpDir := createTempProject(&testing.T{}, "root")
			defer os.RemoveAll(tmpDir)

			// Create specified number of files
			for i := 0; i < count; i++ {
				content := fmt.Sprintf(`package main

import "fmt"

func function%d() {
	fmt.Println("Function %d")
}`, i, i)

				filename := fmt.Sprintf("file%d.go", i)
				if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644); err != nil {
					b.Fatalf("Failed to create file %s: %v", filename, err)
				}
			}

			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				b.Fatalf("Failed to change to temp dir: %v", err)
			}

			app := &WindApp{
				config: WindConfig{
					IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
					ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
				},
				fileStates: make(map[string]time.Time),
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				app.fileStates = make(map[string]time.Time) // Reset state
				if err := app.scanFiles(); err != nil {
					b.Fatalf("Failed to scan files: %v", err)
				}
			}
		})
	}
}

// BenchmarkCompleteWorkflow benchmarks the complete Wind workflow
func BenchmarkCompleteWorkflow(b *testing.B) {
	tmpDir := createTempProject(&testing.T{}, "cmd-api")
	defer os.RemoveAll(tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp dir: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := &WindApp{
			config: WindConfig{
				IncludeExts: []string{".go", ".html", ".css", ".js", ".json", ".yaml", ".yml"},
				ExcludeDirs: []string{"vendor", ".git", "node_modules", "tmp", ".idea", ".vscode"},
			},
			fileStates: make(map[string]time.Time),
		}

		// Simulate complete workflow
		detectProjectStructure()
		app.scanFiles()
		app.checkForChanges()
	}
}

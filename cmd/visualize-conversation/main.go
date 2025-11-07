package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iwamatsu0430/visualize-claude-code-conversation/internal/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var inputFile, outputDir string

	// Parse arguments
	if len(os.Args) >= 3 {
		// Explicit mode: visualize-conversation <input.jsonl> <output-dir>
		inputFile = os.Args[1]
		outputDir = os.Args[2]
	} else {
		// Auto mode: find latest conversation
		var err error
		inputFile, err = findLatestConversation()
		if err != nil {
			return fmt.Errorf("failed to find conversation log: %w\nUsage: %s [input.jsonl output-dir]", err, os.Args[0])
		}

		// Determine output directory: CLI arg > env var > default
		if len(os.Args) >= 2 {
			outputDir = os.Args[1]
		} else if envDir := os.Getenv("VISUALIZE_OUTPUT_DIR"); envDir != "" {
			outputDir = envDir
		} else {
			outputDir = "./dist"
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("Parsing JSONL file...")
	conv, err := parser.ParseJSONL(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse JSONL: %w", err)
	}

	fmt.Printf("Found %d entries\n", len(conv.Entries))

	fmt.Println("Generating HTML...")
	html, err := parser.GenerateHTML(conv)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	outputFile := filepath.Join(outputDir, "index.html")
	fmt.Printf("Writing output to %s...\n", outputFile)

	if err := os.WriteFile(outputFile, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to write HTML: %w", err)
	}

	fmt.Printf("✓ Successfully generated visualization!\n")
	fmt.Printf("  Output: %s\n", outputFile)

	return nil
}

// findLatestConversation finds the latest conversation JSONL file for the current project
func findLatestConversation() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	claudeProjectsDir := filepath.Join(homeDir, ".claude", "projects")

	// Try to find matching project directory
	projectDir, err := findProjectDir(claudeProjectsDir, cwd)
	if err != nil {
		return "", err
	}

	// Find latest JSONL file (excluding agent files)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return "", fmt.Errorf("no conversation logs found for this project (checked: %s)", projectDir)
	}

	var latestFile string
	var latestModTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}

		// Skip agent files
		if strings.HasPrefix(name, "agent-") {
			continue
		}

		fullPath := filepath.Join(projectDir, name)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		modTime := info.ModTime().Unix()
		if modTime > latestModTime {
			latestModTime = modTime
			latestFile = fullPath
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no conversation logs found in %s", projectDir)
	}

	// Extract session ID from filename
	sessionID := strings.TrimSuffix(filepath.Base(latestFile), ".jsonl")
	fmt.Printf("📝 Using session: %s\n", sessionID)

	return latestFile, nil
}

// findProjectDir finds the Claude project directory that matches the current working directory
func findProjectDir(claudeProjectsDir, cwd string) (string, error) {
	entries, err := os.ReadDir(claudeProjectsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read Claude projects directory: %w", err)
	}

	// Normalize current working directory for comparison
	// Remove leading slash and replace separators/dots with dashes
	normalizedCwd := strings.TrimPrefix(cwd, string(os.PathSeparator))
	normalizedCwd = strings.ReplaceAll(normalizedCwd, string(os.PathSeparator), "-")
	normalizedCwd = strings.ReplaceAll(normalizedCwd, ".", "-")

	// Try exact match first
	expectedDirName := "-" + normalizedCwd
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if entry.Name() == expectedDirName {
			return filepath.Join(claudeProjectsDir, entry.Name()), nil
		}
	}

	// Try fuzzy match - find directory that ends with the project name
	projectName := filepath.Base(cwd)
	normalizedProjectName := strings.ReplaceAll(projectName, ".", "-")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), normalizedProjectName) {
			return filepath.Join(claudeProjectsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no Claude project directory found for: %s (tried: %s)", cwd, expectedDirName)
}

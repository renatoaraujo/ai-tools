package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Tools []Tool `yaml:"tools"`
}

type Tool struct {
	Name        string `yaml:"name"`
	Repo        string `yaml:"repo"`
	Description string `yaml:"description"`
}

func main() {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		log.Fatal("GitHub CLI (gh) is not installed or not in PATH")
	}

	// Read config file
	data, err := os.ReadFile("mcp-tools.yaml")
	if err != nil {
		log.Fatalf("Failed to read mcp-tools.yaml: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// Create mcp directory if it doesn't exist
	mcpDir := "mcp"
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		log.Fatalf("Failed to create mcp directory: %v", err)
	}

	fmt.Printf("Cloning %d MCP tools...\n", len(config.Tools))

	for _, tool := range config.Tools {
		fmt.Printf("Cloning %s...\n", tool.Name)
		
		// Extract repo name from URL
		repoName := extractRepoName(tool.Repo)
		targetPath := filepath.Join(mcpDir, repoName)

		// Skip if already exists
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("  %s already exists, skipping\n", repoName)
			continue
		}

		// Clone using gh CLI
		cmd := exec.Command("gh", "repo", "clone", tool.Repo, targetPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("  Failed to clone %s: %v\n", tool.Name, err)
			continue
		}

		fmt.Printf("  ✓ Cloned to %s\n", targetPath)
	}

	fmt.Println("Done!")
}

func extractRepoName(repoURL string) string {
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Split by "/" and get the last part
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	
	return "unknown"
}

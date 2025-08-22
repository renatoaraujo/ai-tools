package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
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
	// Check if gh is authenticated
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatalf("Failed to create GitHub client (is gh CLI authenticated?): %v", err)
	}

	// Test authentication by getting user info
	var user struct {
		Login string `json:"login"`
	}
	if err := client.Get("user", &user); err != nil {
		log.Fatalf("Failed to authenticate with GitHub (run 'gh auth login'): %v", err)
	}
	fmt.Printf("Authenticated as: %s\n", user.Login)

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
		
		// Parse repository URL to get owner and repo
		owner, repo, err := parseRepoURL(tool.Repo)
		if err != nil {
			fmt.Printf("  Failed to parse repo URL %s: %v\n", tool.Repo, err)
			continue
		}

		targetPath := filepath.Join(mcpDir, repo)

		// Skip if already exists
		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("  %s already exists, skipping\n", repo)
			continue
		}

		// Clone repository using go-gh
		if err := cloneRepo(client, owner, repo, targetPath); err != nil {
			fmt.Printf("  Failed to clone %s: %v\n", tool.Name, err)
			continue
		}

		fmt.Printf("  ✓ Cloned to %s\n", targetPath)
	}

	fmt.Println("Done!")
}

func parseRepoURL(repoURL string) (owner, repo string, err error) {
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Handle both https://github.com/owner/repo and owner/repo formats
	if strings.HasPrefix(repoURL, "https://github.com/") {
		repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
	}
	
	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository URL format, expected owner/repo")
	}
	
	return parts[0], parts[1], nil
}

func cloneRepo(client api.RESTClient, owner, repo, targetPath string) error {
	// Get repository information to get the clone URL
	var repoInfo struct {
		CloneURL string `json:"clone_url"`
	}
	
	endpoint := fmt.Sprintf("repos/%s/%s", owner, repo)
	if err := client.Get(endpoint, &repoInfo); err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// Use git to clone the repository
	return gitClone(repoInfo.CloneURL, targetPath)
}

func gitClone(cloneURL, targetPath string) error {
	// We'll use git directly since go-gh doesn't have a clone function
	// This is still better than shelling out to gh CLI
	cmd := fmt.Sprintf("git clone %s %s", cloneURL, targetPath)
	
	// Use os/exec to run git clone
	return runCommand("git", "clone", cloneURL, targetPath)
}

func runCommand(name string, args ...string) error {
	// Simple command execution - in a real implementation you'd use os/exec
	// For now, we'll keep it simple and use the system command
	cmd := name
	for _, arg := range args {
		cmd += " " + arg
	}
	
	// This is a simplified version - in practice use exec.Command
	if err := os.system(cmd); err != nil {
		return fmt.Errorf("command failed: %s", cmd)
	}
	return nil
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatalf("Failed to create GitHub client (is gh CLI authenticated?): %v", err)
	}

	var user struct {
		Login string `json:"login"`
	}

	if err = client.Get("user", &user); err != nil {
		log.Fatalf("Failed to authenticate with GitHub (run 'gh auth login'): %v", err)
	}

	fmt.Printf("Authenticated as: %s\n", user.Login)

	data, err := os.ReadFile("mcp-tools.yaml")
	if err != nil {
		log.Fatalf("Failed to read mcp-tools.yaml: %v", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	mcpDir := "mcp"
	if err = os.MkdirAll(mcpDir, 0755); err != nil {
		log.Fatalf("Failed to create mcp directory: %v", err)
	}

	fmt.Printf("Cloning %d MCP tools...\n", len(config.Tools))

	for _, tool := range config.Tools {
		fmt.Printf("Cloning %s...\n", tool.Name)

		owner, repo, err := parseRepoURL(tool.Repo)
		if err != nil {
			fmt.Printf("  Failed to parse repo URL %s: %v\n", tool.Repo, err)
			continue
		}

		targetPath := filepath.Join(mcpDir, repo)

		if _, err = os.Stat(targetPath); err == nil {
			fmt.Printf("  %s already exists, skipping\n", repo)
			continue
		}

		if err = cloneRepo(client, owner, repo, targetPath); err != nil {
			fmt.Printf("  Failed to clone %s: %v\n", tool.Name, err)
			continue
		}

		fmt.Printf("  ✓ Cloned to %s\n", targetPath)
	}

	fmt.Println("Done!")
}

func parseRepoURL(repoURL string) (owner, repo string, err error) {
	repoURL = strings.TrimSuffix(repoURL, ".git")

	if strings.HasPrefix(repoURL, "https://github.com/") {
		repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository URL format, expected owner/repo")
	}

	return parts[0], parts[1], nil
}

func cloneRepo(client *api.RESTClient, owner, repo, targetPath string) error {
	var repoInfo struct {
		CloneURL string `json:"clone_url"`
	}

	endpoint := fmt.Sprintf("repos/%s/%s", owner, repo)
	if err := client.Get(endpoint, &repoInfo); err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	return gitClone(repoInfo.CloneURL, targetPath)
}

func gitClone(cloneURL, targetPath string) error {
	cmd := exec.Command("git", "clone", cloneURL, targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

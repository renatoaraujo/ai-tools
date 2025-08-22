package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
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

var (
	configFile string
	outputDir  string
	verbose    bool
	version    = "dev" // Will be set during build
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ai-tools",
	Short: "A collection of AI tools and utilities",
	Long:  `ai-tools is a CLI application for managing AI tools and utilities, including cloning MCP (Model Context Protocol) servers from GitHub.`,
	Example: `  # Clone MCP tools using default configuration
  ai-tools clone

  # Clone with custom configuration file
  ai-tools clone --config my-tools.yaml

  # Clone to a custom directory with verbose output
  ai-tools clone --output my-mcp-tools --verbose`,
	// Run clone command by default if no subcommand is specified
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand provided, run clone with default flags
		runClone(cmd, args)
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone MCP tools from GitHub repositories",
	Long:  `Clone MCP (Model Context Protocol) tools from GitHub repositories specified in a configuration file.`,
	Example: `  # Clone using default configuration
  ai-tools clone

  # Clone with custom config and output directory
  ai-tools clone --config my-config.yaml --output ./tools`,
	Run: runClone,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ai-tools",
	Long:  `Print the version number of ai-tools`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ai-tools %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(versionCmd)

	// Add flags to root command for backward compatibility
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "mcp-tools.yaml", "Configuration file containing tools to clone")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "mcp", "Output directory for cloned repositories")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Also add flags to clone command 
	cloneCmd.Flags().StringVarP(&configFile, "config", "c", "mcp-tools.yaml", "Configuration file containing tools to clone")
	cloneCmd.Flags().StringVarP(&outputDir, "output", "o", "mcp", "Output directory for cloned repositories")
	cloneCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func runClone(cmd *cobra.Command, args []string) {
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

	if verbose {
		fmt.Printf("Authenticated as: %s\n", user.Login)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", configFile, err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	if err = os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", outputDir, err)
	}

	fmt.Printf("Cloning %d MCP tools...\n", len(config.Tools))

	for _, tool := range config.Tools {
		fmt.Printf("Cloning %s...\n", tool.Name)

		owner, repo, err := parseRepoURL(tool.Repo)
		if err != nil {
			fmt.Printf("  Failed to parse repo URL %s: %v\n", tool.Repo, err)
			continue
		}

		targetPath := filepath.Join(outputDir, repo)

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

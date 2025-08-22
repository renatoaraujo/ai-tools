# ai-tools
A collection of AI tools and utilities

## Installation

```bash
go install github.com/renatoaraujo/ai-tools@latest
```

Or build from source:

```bash
git clone https://github.com/renatoaraujo/ai-tools.git
cd ai-tools
go build -o ai-tools .
```

## Usage

ai-tools is a CLI application for managing AI tools and utilities, including cloning MCP (Model Context Protocol) servers from GitHub.

### Authentication

Before using ai-tools, make sure you're authenticated with GitHub CLI:

```bash
gh auth login
```

### Basic Usage

```bash
# Clone MCP tools using default configuration (mcp-tools.yaml)
ai-tools

# Or explicitly use the clone command
ai-tools clone

# Clone with custom configuration file
ai-tools clone --config my-tools.yaml

# Clone to a custom directory with verbose output
ai-tools clone --output my-mcp-tools --verbose

# Show version
ai-tools version

# Show help
ai-tools --help
```

### Configuration

Create a `mcp-tools.yaml` file with the tools you want to clone:

```yaml
tools:
  - name: github-mcp-server
    repo: https://github.com/github/github-mcp-server
    description: Official GitHub MCP server
  - name: grafana-mcp-server
    repo: https://github.com/grafana/mcp-grafana
    description: Official Grafana MCP server
```

### Flags

- `--config, -c`: Configuration file containing tools to clone (default: "mcp-tools.yaml")
- `--output, -o`: Output directory for cloned repositories (default: "mcp")
- `--verbose, -v`: Enable verbose output
- `--help, -h`: Show help message

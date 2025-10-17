# Build spacectl CLI Tool

## Overview

Build a Go-based CLI tool using the Cobra framework that interfaces with the Kubespaces backend API. The tool will manage authentication via config file (`~/.spacectl`), support multiple output formats (JSON, YAML, CSV, table), and provide commands for users, organizations, projects, and tenants.

## Project Structure

```
spacectl/
├── cmd/
│   ├── root.go              # Root command and global flags
│   ├── login.go             # Auth: login command
│   ├── register.go          # Auth: register command
│   ├── logout.go            # Auth: logout command
│   ├── whoami.go            # Auth: whoami/user info
│   ├── organization.go      # Org commands (list, create, get, update, delete, set-default)
│   ├── project.go           # Project commands (list, create, get, update, delete)
│   ├── tenant.go            # Tenant commands (list, create, get, delete, status, kubeconfig)
│   └── version.go           # Version command
├── internal/
│   ├── api/
│   │   ├── client.go        # HTTP client with auth handling
│   │   ├── auth.go          # Auth API calls
│   │   ├── organizations.go # Organization API calls
│   │   ├── projects.go      # Project API calls
│   │   └── tenants.go       # Tenant API calls
│   ├── config/
│   │   ├── config.go        # Config file management (~/.spacectl)
│   │   └── auth.go          # Token storage and refresh logic
│   ├── output/
│   │   ├── formatter.go     # Output formatting (JSON, YAML, CSV, table)
│   │   └── table.go         # Table rendering utilities
│   └── models/
│       └── types.go         # Shared types matching backend models
├── go.mod
├── go.sum
├── main.go                  # Entry point
├── Makefile                 # Build commands
└── README.md                # CLI documentation
```

## Key Features

### 1. Authentication & Config

- Store access/refresh tokens in `~/.spacectl` (JSON format)
- Auto-refresh tokens when expired
- Support `spacectl login`, `spacectl logout`, `spacectl whoami`
- Config structure:
  ```json
  {
    "api_url": "http://localhost:8080",
    "access_token": "...",
    "refresh_token": "...",
    "user_email": "..."
  }
  ```


### 2. Global Flags

- `--output, -o`: Output format (json, yaml, csv, table) - default: table
- `--api-url`: Override API URL from config
- `--no-headers`: Suppress headers in table/CSV output
- `--quiet, -q`: Minimal output

### 3. Command Structure

**Authentication:**

- `spacectl login [--email] [--password]` - Interactive if flags omitted
- `spacectl register [--email] [--password]`
- `spacectl logout` - Clear stored credentials
- `spacectl whoami` - Display current user info

**Organizations:**

- `spacectl org list` - List user's organizations
- `spacectl org create <name>` - Create organization
- `spacectl org get <id>` - Get organization details
- `spacectl org update <id> --name <name>` - Update organization
- `spacectl org delete <id>` - Delete organization
- `spacectl org set-default <id>` - Set default organization
- `spacectl org members list <id>` - List organization members
- `spacectl org members add <org-id> --user <user-id> --role <role>`
- `spacectl org members remove <org-id> <user-id>`

**Projects:**

- `spacectl project list [--org <id>]` - List projects
- `spacectl project create <name> --org <id> [--description] [--max-tenants] [--max-compute] [--max-memory]`
- `spacectl project get <id>` - Get project details
- `spacectl project update <id> [--name] [--description]`
- `spacectl project delete <id>`
- `spacectl project members list <id>`
- `spacectl project members add <project-id> --user <user-id> --role <role>`

**Tenants:**

- `spacectl tenant list [--project <id>]` - List tenants
- `spacectl tenant create <name> --project <id> --cloud <provider> --region <region> --k8s-version <version> --compute <cores> --memory <gb>`
- `spacectl tenant get <id>` - Get tenant details
- `spacectl tenant delete <id>`
- `spacectl tenant status <id>` - Get provisioning status
- `spacectl tenant kubeconfig <id> [--output-file <path>]` - Download kubeconfig
- `spacectl tenant locations` - List available locations
- `spacectl tenant k8s-versions` - List available Kubernetes versions

### 4. Output Formatting

- **Table**: Human-readable ASCII tables using a library like `tablewriter`
- **JSON**: Pretty-printed JSON
- **YAML**: YAML format
- **CSV**: Comma-separated values with headers

### 5. Error Handling

- Friendly error messages for common errors (401, 403, 404, etc.)
- Suggest `spacectl login` when auth fails
- Auto-refresh tokens on 401 if refresh token available

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Config management (optional, can use encoding/json)
- `gopkg.in/yaml.v3` - YAML output
- `github.com/olekukonko/tablewriter` - Table formatting
- Standard library for JSON/CSV

## Implementation Order

1. Setup project structure and dependencies
2. Implement config management (read/write ~/.spacectl)
3. Build API client with auth handling
4. Implement auth commands (login, logout, whoami)
5. Add organization commands
6. Add project commands
7. Add tenant commands
8. Implement output formatters
9. Add Makefile and documentation
10. Test all commands end-to-end
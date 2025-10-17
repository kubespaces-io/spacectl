# spacectl

A command-line tool for managing Kubespaces resources including organizations, projects, and tenants.

It will be available as a brew package.

## Features

- **Authentication**: Login, register, and manage authentication tokens
- **Organizations**: Create, list, update, and delete organizations
- **Projects**: Manage projects within organizations
- **Tenants**: Create and manage Kubernetes tenants
- **Multiple Output Formats**: JSON, YAML, CSV, and human-readable tables
- **Auto-refresh Tokens**: Automatic token refresh when expired

## Installation

### From Source

```bash
git clone <repository-url>
cd spacectl
make build
sudo cp bin/spacectl /usr/local/bin/
```

### Development Setup

```bash
make dev-setup
make build
```

### Shell Autocompletion

spacectl supports shell autocompletion for zsh and bash. To set it up:

#### Automatic Setup (Recommended)

```bash
# Run the setup script
./scripts/setup-completion.sh
```

#### Manual Setup

**For zsh:**

```bash
# Install completion
spacectl completion zsh > $(brew --prefix)/share/zsh/site-functions/_spacectl

# Enable completion in your shell
echo "autoload -U compinit; compinit" >> ~/.zshrc
source ~/.zshrc
```

**For bash:**

```bash
# Install completion
spacectl completion bash > /usr/local/etc/bash_completion.d/spacectl

# Enable completion in your shell
echo "source /usr/local/etc/bash_completion.d/spacectl" >> ~/.bashrc
source ~/.bashrc
```

After setup, you can use `<TAB>` to autocomplete commands, flags, and options.

## Configuration

spacectl stores configuration in `~/.spacectl`:

```json
{
  "api_url": "http://localhost:8080",
  "access_token": "...",
  "refresh_token": "...",
  "user_email": "..."
}
```

## Usage

### Authentication

```bash
# Login with email/password interactively
spacectl auth login

# Login with email/password using flags
spacectl auth login --email user@example.com --password mypassword

# Login with GitHub OAuth (opens browser)
spacectl auth github-login

# Register a new account
spacectl register --email user@example.com --password mypassword

# Check current user
spacectl whoami

# Logout
spacectl logout
```

### Organizations

```bash
# List organizations
spacectl org list

# Create organization
spacectl org create "My Organization"

# Get organization details
spacectl org get <org-id>

# Update organization
spacectl org update <org-id> --name "New Name"

# Set default organization
spacectl org set-default <org-id>

# Delete organization
spacectl org delete <org-id>
```

### Projects

```bash
# List projects
spacectl project list

# List projects in organization
spacectl project list --org <org-id>

# Create project
spacectl project create "My Project" --org <org-id> --description "Project description"

# Get project details
spacectl project get <project-id>

# Update project
spacectl project update <project-id> --name "New Name" --description "New description"

# Delete project
spacectl project delete <project-id>

# Manage project members
spacectl project members list <project-id>
spacectl project members add <project-id> --user <user-id> --role admin
spacectl project members remove <project-id> <user-id>
```

### Tenants

```bash
# List tenants
spacectl tenant list --project <project-id>

# Create tenant
spacectl tenant create "my-tenant" \
  --project <project-id> \
  --cloud gcp \
  --region us-central1 \
  --k8s-version v1.33.1 \
  --compute 2 \
  --memory 4

# Get tenant details
spacectl tenant get <tenant-id>

# Get tenant status
spacectl tenant status <tenant-id>

# Download kubeconfig
spacectl tenant kubeconfig <tenant-id> --output-file ~/.kube/config

# Delete tenant
spacectl tenant delete <tenant-id>

# List available locations
spacectl tenant locations

# List available Kubernetes versions
spacectl tenant k8s-versions
```

### Output Formats

```bash
# Table format (default)
spacectl org list

# JSON format
spacectl org list --output json

# YAML format
spacectl org list --output yaml

# CSV format
spacectl org list --output csv

# Suppress headers
spacectl org list --output csv --no-headers

# Quiet mode
spacectl org create "My Org" --quiet
```

### Global Flags

- `--api-url`: Override API URL from config
- `--output, -o`: Output format (table, json, yaml, csv)
- `--no-headers`: Suppress headers in table/CSV output
- `--quiet, -q`: Minimal output

## Examples

### Complete Workflow

```bash
# 1. Login
spacectl login

# 2. Create organization
spacectl org create "My Company"

# 3. Create project
spacectl project create "Web App" --org <org-id> --description "Main web application"

# 4. Create tenant
spacectl tenant create "dev-cluster" \
  --project <project-id> \
  --cloud gcp \
  --region us-central1 \
  --k8s-version v1.33.1 \
  --compute 2 \
  --memory 4

# 5. Download kubeconfig
spacectl tenant kubeconfig <tenant-id> --output-file ~/.kube/config

# 6. Use with kubectl
kubectl get nodes
```

### Using with kubectl

```bash
# Download kubeconfig for a tenant
spacectl tenant kubeconfig <tenant-id> --output-file ~/.kube/tenant-config

# Use with kubectl
kubectl --kubeconfig ~/.kube/tenant-config get pods

# Or set as default
export KUBECONFIG=~/.kube/tenant-config
kubectl get pods
```

## Development

### Building

```bash
make build          # Build binary
make build-all      # Build for all platforms
make install        # Install to $GOPATH/bin
```

### Testing

```bash
make test           # Run tests
make lint           # Run linter
```

### Running

```bash
make run ARGS="--help"  # Run with arguments
```

## API Compatibility

spacectl is compatible with the Kubespaces API v1. It communicates with the backend using the following endpoints:

- Authentication: `/api/v1/user/*`
- Organizations: `/api/v1/organizations/*`
- Projects: `/api/v1/projects/*`
- Tenants: `/api/v1/tenants/*`

## Error Handling

spacectl provides friendly error messages for common scenarios:

- **401 Unauthorized**: Suggests running `spacectl login`
- **403 Forbidden**: Indicates insufficient permissions
- **404 Not Found**: Resource doesn't exist
- **Network errors**: Connection issues with the API

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make lint`
6. Submit a pull request

## License

MIT License

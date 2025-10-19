#!/bin/bash

# Setup script for spacectl autocompletion
# This script sets up shell completion for spacectl

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up spacectl autocompletion...${NC}"

# Check if spacectl is available
if ! command -v spacectl &> /dev/null; then
    echo -e "${RED}Error: spacectl not found in PATH${NC}"
    echo "Please install spacectl first:"
    echo "  make build && sudo cp bin/spacectl /usr/local/bin/"
    exit 1
fi

# Detect shell
SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    "zsh")
        echo -e "${YELLOW}Setting up zsh completion...${NC}"
        
        # Check if Homebrew is available
        if command -v brew &> /dev/null; then
            COMPLETION_DIR="$(brew --prefix)/share/zsh/site-functions"
        else
            COMPLETION_DIR="/usr/local/share/zsh/site-functions"
        fi
        
        # Create directory if it doesn't exist
        mkdir -p "$COMPLETION_DIR"
        
        # Generate and install completion
        spacectl completion zsh > "$COMPLETION_DIR/_spacectl"
        
        echo -e "${GREEN}✓ Zsh completion installed to: $COMPLETION_DIR/_spacectl${NC}"
        echo ""
        echo "To enable completion:"
        echo "  1. Add to ~/.zshrc:"
        echo "     echo 'autoload -U compinit; compinit' >> ~/.zshrc"
        echo ""
        echo "  2. Reload your shell:"
        echo "     source ~/.zshrc"
        echo ""
        echo "  3. Or reload completions immediately:"
        echo "     autoload -U compinit && compinit"
        ;;
        
    "bash")
        echo -e "${YELLOW}Setting up bash completion...${NC}"
        
        # Check for bash completion directory
        if [ -d "/usr/local/etc/bash_completion.d" ]; then
            COMPLETION_DIR="/usr/local/etc/bash_completion.d"
        elif [ -d "/etc/bash_completion.d" ]; then
            COMPLETION_DIR="/etc/bash_completion.d"
        else
            COMPLETION_DIR="$HOME/.local/share/bash-completion/completions"
            mkdir -p "$COMPLETION_DIR"
        fi
        
        # Generate and install completion
        spacectl completion bash > "$COMPLETION_DIR/spacectl"
        
        echo -e "${GREEN}✓ Bash completion installed to: $COMPLETION_DIR/spacectl${NC}"
        echo ""
        echo "To enable completion:"
        echo "  1. Add to ~/.bashrc:"
        echo "     echo 'source $COMPLETION_DIR/spacectl' >> ~/.bashrc"
        echo ""
        echo "  2. Reload your shell:"
        echo "     source ~/.bashrc"
        ;;
        
    *)
        echo -e "${YELLOW}Unsupported shell: $SHELL_NAME${NC}"
        echo "Supported shells: zsh, bash"
        echo ""
        echo "You can manually generate completion:"
        echo "  spacectl completion zsh > /path/to/completion/file"
        echo "  spacectl completion bash > /path/to/completion/file"
        exit 1
        ;;
esac

echo -e "${GREEN}✓ Autocompletion setup complete!${NC}"
echo ""
echo "Test it by typing:"
echo "  spacectl <TAB><TAB>"
echo ""
echo "You should see available commands and options."




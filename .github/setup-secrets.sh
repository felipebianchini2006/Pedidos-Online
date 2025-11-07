#!/bin/bash

################################################################################
# GitHub Actions Secrets Setup Script
################################################################################
# This script helps you configure GitHub secrets for CI/CD
# Usage: bash .github/setup-secrets.sh
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_header() {
    echo -e "\n${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

################################################################################
# Main
################################################################################

print_header "🔐 GitHub Actions Secrets Setup"

echo "This script will help you configure the necessary secrets for CI/CD."
echo ""

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
    print_error "GitHub CLI (gh) is not installed"
    echo ""
    echo "Install it with:"
    echo "  macOS:   brew install gh"
    echo "  Linux:   see https://github.com/cli/cli/blob/trunk/docs/install_linux.md"
    echo "  Windows: see https://github.com/cli/cli/releases"
    echo ""
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    print_warning "Not authenticated with GitHub CLI"
    echo ""
    print_info "Authenticating..."
    gh auth login
fi

print_success "GitHub CLI is authenticated"

# Get repository info
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
print_info "Repository: $REPO"

echo ""
print_header "📋 Secret Configuration"

################################################################################
# Docker Hub (Optional)
################################################################################

echo ""
print_info "1. Docker Hub Configuration (Optional)"
echo ""
read -p "Do you want to configure Docker Hub? (y/N): " configure_docker

if [[ $configure_docker =~ ^[Yy]$ ]]; then
    read -p "Docker Hub Username: " docker_username
    read -sp "Docker Hub Password/Token: " docker_password
    echo ""
    
    gh secret set DOCKER_USERNAME -b "$docker_username"
    gh secret set DOCKER_PASSWORD -b "$docker_password"
    
    print_success "Docker Hub secrets configured"
else
    print_info "Skipping Docker Hub (will use GitHub Container Registry)"
fi

################################################################################
# Production Server
################################################################################

echo ""
print_info "2. Production Server Configuration"
echo ""

# SSH Host
read -p "Production SSH Host (IP or hostname): " ssh_host
gh secret set PROD_SSH_HOST -b "$ssh_host"

# SSH User
read -p "Production SSH User: " ssh_user
gh secret set PROD_SSH_USER -b "$ssh_user"

# SSH Key
echo ""
print_info "SSH Key Configuration"
echo ""
echo "Choose an option:"
echo "  1. Generate new SSH key"
echo "  2. Use existing SSH key"
echo ""
read -p "Option (1/2): " ssh_option

if [ "$ssh_option" == "1" ]; then
    # Generate new key
    KEY_PATH="$HOME/.ssh/github_actions_deploy"
    
    if [ -f "$KEY_PATH" ]; then
        print_warning "Key already exists at $KEY_PATH"
        read -p "Overwrite? (y/N): " overwrite
        if [[ ! $overwrite =~ ^[Yy]$ ]]; then
            print_info "Using existing key"
        else
            ssh-keygen -t ed25519 -C "github-actions-deploy" -f "$KEY_PATH" -N ""
            print_success "New SSH key generated"
        fi
    else
        ssh-keygen -t ed25519 -C "github-actions-deploy" -f "$KEY_PATH" -N ""
        print_success "SSH key generated at $KEY_PATH"
    fi
    
    # Set secret
    gh secret set PROD_SSH_KEY < "$KEY_PATH"
    print_success "SSH private key configured as secret"
    
    # Display public key
    echo ""
    print_info "Add this public key to your server:"
    echo ""
    cat "$KEY_PATH.pub"
    echo ""
    print_warning "Run this on your server:"
    echo "  cat >> ~/.ssh/authorized_keys << 'EOF'"
    cat "$KEY_PATH.pub"
    echo "  EOF"
    echo ""
    read -p "Press Enter after adding the public key to your server..."
    
elif [ "$ssh_option" == "2" ]; then
    # Use existing key
    read -p "Path to SSH private key: " key_path
    
    if [ ! -f "$key_path" ]; then
        print_error "File not found: $key_path"
        exit 1
    fi
    
    gh secret set PROD_SSH_KEY < "$key_path"
    print_success "SSH private key configured as secret"
fi

################################################################################
# Test SSH Connection
################################################################################

echo ""
print_info "Testing SSH connection..."
echo ""

if ssh -i "$KEY_PATH" -o StrictHostKeyChecking=no -o ConnectTimeout=5 "$ssh_user@$ssh_host" "echo 'Connection successful'" 2>/dev/null; then
    print_success "SSH connection successful!"
else
    print_error "SSH connection failed"
    print_warning "Please verify:"
    echo "  1. Public key is added to server's ~/.ssh/authorized_keys"
    echo "  2. Server is reachable from your current network"
    echo "  3. SSH user and host are correct"
fi

################################################################################
# Codecov (Optional)
################################################################################

echo ""
print_info "3. Codecov Configuration (Optional)"
echo ""
read -p "Do you want to configure Codecov? (y/N): " configure_codecov

if [[ $configure_codecov =~ ^[Yy]$ ]]; then
    echo ""
    echo "Get your Codecov token from:"
    echo "  https://codecov.io/gh/$REPO"
    echo ""
    read -sp "Codecov Token: " codecov_token
    echo ""
    
    gh secret set CODECOV_TOKEN -b "$codecov_token"
    print_success "Codecov token configured"
else
    print_info "Skipping Codecov"
fi

################################################################################
# Summary
################################################################################

print_header "✅ Configuration Complete!"

echo "Secrets configured:"
echo ""

gh secret list

echo ""
print_info "Next steps:"
echo "  1. Verify secrets in GitHub: https://github.com/$REPO/settings/secrets/actions"
echo "  2. Test CI/CD by creating a pull request"
echo "  3. Monitor workflow runs: https://github.com/$REPO/actions"
echo ""

print_success "Setup completed successfully!"

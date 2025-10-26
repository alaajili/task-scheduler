#!/bin/bash

set -e

echo "🚀 Setting up Task Scheduler development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        return 1
    else
        echo -e "${GREEN}✓ $1 is installed${NC}"
        return 0
    fi
}

# --- Install golang-migrate ---
install_migrate() {
    if check_command migrate; then
        echo -e "${GREEN}migrate is already installed${NC}"
        return
    fi

    echo -e "${YELLOW}Installing golang-migrate...${NC}"
    OS=$(uname -s)

    case "$OS" in
        Darwin)
            if check_command brew; then
                brew install golang-migrate
            else
                echo -e "${RED}Homebrew not found.${NC}"
                echo "Please install Homebrew first: https://brew.sh"
                exit 1
            fi
            ;;
        Linux)
            if check_command apt; then
                sudo apt update && sudo apt install -y golang-migrate
            elif check_command dnf; then
                sudo dnf install -y golang-migrate
            else
                echo -e "${RED}Unsupported Linux distribution.${NC}"
                echo "Install manually from: https://github.com/golang-migrate/migrate/releases"
                exit 1
            fi
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo -e "${RED}Detected Windows environment.${NC}"
            echo "Please install golang-migrate manually from:"
            echo "https://github.com/golang-migrate/migrate/releases"
            exit 1
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
}

echo "Checking prerequisites..."
check_command go || exit 1
check_command docker || exit 1
check_command docker-compose || exit 1
check_command make || exit 1

# Install migrate if not present
if ! check_command migrate; then
    echo -e "${YELLOW}Installing golang-migrate...${NC}"
    install_migrate
else
    echo -e "${GREEN}migrate is already installed${NC}"
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p bin logs

# Copy example config
if [ ! -f config.yaml ]; then
    echo "Creating config.yaml from example..."
    cp config.example.yaml config.yaml
fi

# Install Go dependencies
echo "Installing Go dependencies..."
make setup

# Start Docker services
echo "Starting Docker services..."
make docker-up

# Wait for services to be healthy
echo "Waiting for services to be ready..."
sleep 10

# Run migrations
echo "Running database migrations..."
make migrate-up

echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Run 'make run-api' to start the API server"
echo "  2. Run 'make run-scheduler' to start the scheduler"
echo "  3. Run 'make run-worker' to start a worker"
echo ""
echo "Or run 'make help' to see all available commands"

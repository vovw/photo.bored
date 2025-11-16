#!/bin/bash
set -e

echo "🚀 Photo.Bored Deployment Script"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root for system deployment
if [ "$1" == "system" ]; then
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run as root for system deployment${NC}"
        exit 1
    fi
fi

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Deployment type
DEPLOY_TYPE=${1:-docker}

case $DEPLOY_TYPE in
    docker)
        print_info "Deploying with Docker..."

        # Check if docker is installed
        if ! command -v docker &> /dev/null; then
            print_error "Docker is not installed. Please install Docker first."
            exit 1
        fi

        # Build and start
        print_info "Building Docker image..."
        docker-compose build

        print_info "Starting services..."
        docker-compose up -d

        print_info "Checking health..."
        sleep 5
        if docker-compose ps | grep -q "Up"; then
            print_info "✅ Deployment successful!"
            print_info "Access your app at http://localhost:8080"
        else
            print_error "Deployment failed. Check logs with: docker-compose logs"
            exit 1
        fi
        ;;

    system)
        print_info "Deploying as system service..."

        # Build the application
        print_info "Building application..."
        go build -o app .

        # Create installation directory
        INSTALL_DIR="/opt/photo-bored"
        print_info "Installing to $INSTALL_DIR..."

        mkdir -p $INSTALL_DIR
        mkdir -p $INSTALL_DIR/data
        mkdir -p $INSTALL_DIR/uploads

        # Copy files
        cp app $INSTALL_DIR/
        cp index.html $INSTALL_DIR/

        # Set permissions
        chown -R www-data:www-data $INSTALL_DIR
        chmod +x $INSTALL_DIR/app

        # Install systemd service
        print_info "Installing systemd service..."
        cp photo-bored.service /etc/systemd/system/

        systemctl daemon-reload
        systemctl enable photo-bored
        systemctl restart photo-bored

        print_info "Checking service status..."
        sleep 2
        if systemctl is-active --quiet photo-bored; then
            print_info "✅ Deployment successful!"
            print_info "Service is running. Check status with: systemctl status photo-bored"
        else
            print_error "Service failed to start. Check logs with: journalctl -u photo-bored -f"
            exit 1
        fi
        ;;

    local)
        print_info "Building for local development..."

        # Create directories
        mkdir -p data uploads

        # Build
        print_info "Building application..."
        go build -o app .

        print_info "✅ Build successful!"
        print_info "Run with: ./app"
        print_info "Or with custom config: PORT=3000 ./app"
        ;;

    *)
        print_error "Unknown deployment type: $DEPLOY_TYPE"
        echo "Usage: $0 [docker|system|local]"
        echo ""
        echo "Options:"
        echo "  docker  - Deploy using Docker Compose (default)"
        echo "  system  - Deploy as systemd service (requires root)"
        echo "  local   - Build for local development"
        exit 1
        ;;
esac

echo ""
print_info "Deployment complete! 🎉"

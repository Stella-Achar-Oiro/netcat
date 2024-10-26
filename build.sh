#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Project name
PROJECT="TCPChat"

# Function to print step information
print_step() {
    echo -e "${YELLOW}Step: $1${NC}"
}

# Function to check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        exit 1
    fi
}

# Function to display help
show_help() {
    echo "Usage: ./build.sh [option]"
    echo "Options:"
    echo "  --help          Show this help message"
    echo "  --clean         Clean the project"
    echo "  --test          Run tests only"
    echo "  --build         Build for current platform"
    echo "  --all          Build for all platforms"
    echo "  --run          Build and run the server"
    echo "  --run-ui       Build and run with UI"
    echo
    echo "Examples:"
    echo "  ./build.sh --clean"
    echo "  ./build.sh --build"
    echo "  ./build.sh --run 8989"
}

# Function to clean
clean() {
    print_step "Cleaning project"
    rm -rf bin
    rm -f $PROJECT
    rm -f *.log
    go clean
    echo -e "${GREEN}Clean complete${NC}"
}

# Function to get dependencies
get_deps() {
    print_step "Getting dependencies"
    go mod init $PROJECT 2>/dev/null || true
    if ! go get github.com/jroimartin/gocui; then
        echo -e "${RED}Failed to get gocui package${NC}"
        exit 1
    fi
    go mod tidy
    echo -e "${GREEN}Dependencies installed${NC}"
}

# Function to run tests
run_tests() {
    print_step "Running tests"
    if ! go test -v ./...; then
        echo -e "${RED}Tests failed${NC}"
        exit 1
    fi
    echo -e "${GREEN}Tests passed${NC}"
}

# Function to build
build() {
    print_step "Building project"
    if ! go build -o $PROJECT *.go; then
        echo -e "${RED}Build failed${NC}"
        exit 1
    fi
    chmod +x $PROJECT
    echo -e "${GREEN}Build complete${NC}"
}

# Function to build for all platforms
build_all() {
    print_step "Building for multiple platforms"
    
    # Create bin directory
    mkdir -p bin
    
    # Platforms to build for
    declare -A platforms
    platforms["windows/amd64"]=".exe"
    platforms["windows/386"]=".exe"
    platforms["linux/amd64"]=""
    platforms["linux/386"]=""
    platforms["darwin/amd64"]=""

    # Build for each platform
    for platform in "${!platforms[@]}"; do
        IFS='/' read -r -a parts <<< "$platform"
        GOOS="${parts[0]}"
        GOARCH="${parts[1]}"
        suffix="${platforms[$platform]}"
        
        echo "Building for $GOOS/$GOARCH..."
        output="bin/$PROJECT-$GOOS-$GOARCH$suffix"
        
        if GOOS=$GOOS GOARCH=$GOARCH go build -o "$output" *.go; then
            echo -e "${GREEN}Built $output${NC}"
        else
            echo -e "${RED}Failed to build for $GOOS/$GOARCH${NC}"
        fi
    done
    
    echo -e "${GREEN}Multi-platform builds complete${NC}"
}

# Function to run the server
run_server() {
    port=${1:-8989}
    print_step "Running server on port $port"
    ./$PROJECT "$port"
}

# Function to run the server with UI
run_server_ui() {
    port=${1:-8989}
    print_step "Running server with UI on port $port"
    ./$PROJECT -ui "$port"
}

# Check for required commands
check_command "go"
check_command "git"

# Process command line arguments
case "$1" in
    --help)
        show_help
        ;;
    --clean)
        clean
        ;;
    --test)
        get_deps
        run_tests
        ;;
    --build)
        get_deps
        run_tests
        build
        ;;
    --all)
        clean
        get_deps
        run_tests
        build_all
        ;;
    --run)
        get_deps
        run_tests
        build
        run_server "$2"
        ;;
    --run-ui)
        get_deps
        run_tests
        build
        run_server_ui "$2"
        ;;
    "")
        # Default: build for current platform
        get_deps
        run_tests
        build
        ;;
    *)
        echo -e "${RED}Unknown option: $1${NC}"
        show_help
        exit 1
        ;;
esac

# Print final instructions if build was successful
if [ -f "$PROJECT" ]; then
    echo
    echo "Build successful! You can run the server using:"
    echo "  ./$PROJECT [port]     - Run in normal mode"
    echo "  ./$PROJECT -ui [port] - Run with Terminal UI"
    echo
    echo "Default port is 8989 if not specified"
fi
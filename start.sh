#!/bin/bash

# GPU Platform - Quick Start Script

set -e

echo "========================================"
echo "GPU Platform - Quick Start"
echo "========================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}Warning: Node.js is not installed (frontend build will be skipped)${NC}"
    fi
    
    echo -e "${GREEN}Prerequisites check passed${NC}"
}

# Build backend
build_backend() {
    echo ""
    echo "Building backend..."
    cd /workspace
    mkdir -p bin
    go build -o bin/api-server ./cmd/api-server/
    echo -e "${GREEN}Backend built successfully${NC}"
}

# Build frontend
build_frontend() {
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}Skipping frontend build (Node.js not installed)${NC}"
        return
    fi
    
    echo ""
    echo "Building frontend..."
    cd /workspace/frontend
    npm install
    npm run build
    echo -e "${GREEN}Frontend built successfully${NC}"
}

# Start services
start_services() {
    echo ""
    echo "Starting services..."
    
    # Start backend in background
    cd /workspace
    ./bin/api-server &
    BACKEND_PID=$!
    echo "Backend started (PID: $BACKEND_PID)"
    
    # Start frontend dev server in background
    cd /workspace/frontend
    npm run dev &
    FRONTEND_PID=$!
    echo "Frontend started (PID: $FRONTEND_PID)"
    
    echo ""
    echo "========================================"
    echo -e "${GREEN}GPU Platform is running!${NC}"
    echo "========================================"
    echo ""
    echo "Access URLs:"
    echo "  - Frontend: http://localhost:3000"
    echo "  - Backend API: http://localhost:8080"
    echo "  - Health Check: http://localhost:8080/health"
    echo ""
    echo "Press Ctrl+C to stop all services"
    
    # Wait for user interrupt
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT
    wait
}

# Main
main() {
    check_prerequisites
    build_backend
    build_frontend
    start_services
}

main "$@"

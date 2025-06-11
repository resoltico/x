#!/bin/bash

# Engraving Processor Pro - Startup Script with Checks
# This script ensures everything is ready before starting the server
# 
# Make this script executable with: chmod +x start-with-checks.sh

echo "🚀 Engraving Processor Pro - Starting up..."
echo "================================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check Node.js version
echo -e "\n1️⃣  Checking Node.js version..."
NODE_VERSION=$(node -v)
NODE_MAJOR=$(echo $NODE_VERSION | cut -d. -f1 | sed 's/v//')
if [ $NODE_MAJOR -ge 22 ]; then
    echo -e "${GREEN}✅ Node.js $NODE_VERSION${NC}"
else
    echo -e "${RED}❌ Node.js $NODE_VERSION is too old (need v22+)${NC}"
    exit 1
fi

# Check if build exists
echo -e "\n2️⃣  Checking build..."
if [ -d "build" ] && [ -f "build/server/index.js" ]; then
    echo -e "${GREEN}✅ Build found${NC}"
else
    echo -e "${YELLOW}⚠️  Build not found, running build...${NC}"
    npm run build
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
    fi
fi

# Run startup checks
echo -e "\n3️⃣  Running startup checks..."
node startup-checks.js
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Startup checks failed${NC}"
    exit 1
fi

# Kill any existing process on port 3000
PORT=${PORT:-3000}
echo -e "\n4️⃣  Checking port $PORT..."
if lsof -i :$PORT > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Port $PORT is in use, attempting to free it...${NC}"
    PID=$(lsof -t -i :$PORT)
    if [ ! -z "$PID" ]; then
        kill -9 $PID
        sleep 1
        echo -e "${GREEN}✅ Port $PORT freed${NC}"
    fi
else
    echo -e "${GREEN}✅ Port $PORT is available${NC}"
fi

# Start the server
echo -e "\n5️⃣  Starting server..."
echo "================================================"
npm start &
SERVER_PID=$!

# Wait for server to start
echo -e "\n⏳ Waiting for server to start..."
sleep 3

# Run health check
echo -e "\n6️⃣  Running health check..."
node test-server-health.js
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}🎉 Server is running successfully!${NC}"
    echo -e "\n📌 Server URLs:"
    echo -e "   Web UI:     http://localhost:$PORT"
    echo -e "   Health:     http://localhost:$PORT/health"
    echo -e "   WebSocket:  ws://localhost:$PORT/ws"
    echo -e "\n💡 Press Ctrl+C to stop the server"
    
    # Keep the server running
    wait $SERVER_PID
else
    echo -e "\n${RED}❌ Health check failed${NC}"
    kill $SERVER_PID
    exit 1
fi
#!/bin/bash

# Rate Limiter Load Test Script

echo "🚀 Starting Rate Limiter Load Tests..."
echo "======================================"

# Configuration
BASE_URL="http://localhost:8080"
ENDPOINTS=("/health" "/api/test" "/api/data")
TOKENS=("" "abc123" "vip_token")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to make HTTP request and show result
make_request() {
    local endpoint=$1
    local token=$2
    local token_header=""
    
    if [ ! -z "$token" ]; then
        token_header="-H \"API_KEY: $token\""
    fi
    
    response=$(eval "curl -s -w \"HTTPSTATUS:%{http_code};HEADERS:%{header_json}\" $token_header $BASE_URL$endpoint")
    
    # Extract status code and headers
    http_status=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    headers=$(echo "$response" | grep -o "HEADERS:{.*}" | cut -d: -f2-)
    body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*;HEADERS:{.*}//')
    
    # Extract rate limit headers
    rate_limit=$(echo "$headers" | grep -o '"x-ratelimit-limit":"[^"]*"' | cut -d: -f2 | tr -d '"')
    remaining=$(echo "$headers" | grep -o '"x-ratelimit-remaining":"[^"]*"' | cut -d: -f2 | tr -d '"')
    reset=$(echo "$headers" | grep -o '"x-ratelimit-reset":"[^"]*"' | cut -d: -f2 | tr -d '"')
    limit_type=$(echo "$headers" | grep -o '"x-ratelimit-type":"[^"]*"' | cut -d: -f2 | tr -d '"')
    
    # Color code based on status
    if [ "$http_status" -eq 200 ]; then
        status_color=$GREEN
    elif [ "$http_status" -eq 429 ]; then
        status_color=$RED
    else
        status_color=$YELLOW
    fi
    
    printf "${status_color}%s${NC} | Token: %-10s | Remaining: %-2s | Type: %-5s | Endpoint: %s\n" \
           "$http_status" "${token:-none}" "${remaining:-N/A}" "${limit_type:-N/A}" "$endpoint"
}

# Function to test IP rate limiting
test_ip_limiting() {
    echo -e "\n${BLUE}📊 Testing IP Rate Limiting${NC}"
    echo "Expected: 10 requests allowed per second per IP"
    echo "--------------------------------------------"
    
    for i in {1..15}; do
        make_request "/api/test" ""
        sleep 0.1
    done
}

# Function to test token rate limiting
test_token_limiting() {
    echo -e "\n${BLUE}🔑 Testing Token Rate Limiting${NC}"
    echo "Expected: Higher limits for tokens (if configured)"
    echo "------------------------------------------------"
    
    for token in "${TOKENS[@]}"; do
        if [ ! -z "$token" ]; then
            echo -e "\n${YELLOW}Testing with token: $token${NC}"
            for i in {1..12}; do
                make_request "/api/test" "$token"
                sleep 0.1
            done
        fi
    done
}

# Function to test different endpoints
test_endpoints() {
    echo -e "\n${BLUE}🎯 Testing Different Endpoints${NC}"
    echo "All endpoints share the same rate limit"
    echo "-------------------------------------"
    
    for endpoint in "${ENDPOINTS[@]}"; do
        echo -e "\n${YELLOW}Testing endpoint: $endpoint${NC}"
        for i in {1..5}; do
            make_request "$endpoint" ""
            sleep 0.2
        done
    done
}

# Function to test concurrent requests
test_concurrent() {
    echo -e "\n${BLUE}⚡ Testing Concurrent Requests${NC}"
    echo "Simulating multiple concurrent users"
    echo "-----------------------------------"
    
    # Launch multiple background processes
    for i in {1..5}; do
        {
            for j in {1..3}; do
                make_request "/api/data" "" 
                sleep 0.1
            done
        } &
    done
    
    # Wait for all background processes to complete
    wait
}

# Function to test recovery after rate limit
test_recovery() {
    echo -e "\n${BLUE}🔄 Testing Rate Limit Recovery${NC}"
    echo "Testing that limits reset after block time"
    echo "----------------------------------------"
    
    # Exceed the limit
    echo "Exceeding rate limit..."
    for i in {1..15}; do
        make_request "/health" ""
    done
    
    echo -e "\n${YELLOW}Waiting for rate limit to reset (30 seconds)...${NC}"
    sleep 30
    
    echo "Testing after reset:"
    for i in {1..3}; do
        make_request "/health" ""
        sleep 1
    done
}

# Function to check server health
check_server() {
    echo -e "\n${BLUE}🏥 Checking Server Health${NC}"
    echo "----------------------------"
    
    health_response=$(curl -s "$BASE_URL/health")
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Server is running${NC}"
        echo "Response: $health_response"
        return 0
    else
        echo -e "${RED}✗ Server is not responding${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo "Testing Rate Limiter at: $BASE_URL"
    echo "Time: $(date)"
    
    # Check if server is running
    if ! check_server; then
        echo -e "\n${RED}❌ Cannot proceed - server is not running${NC}"
        echo "Please start the server with: docker-compose up"
        exit 1
    fi
    
    # Run tests
    test_ip_limiting
    test_token_limiting
    test_endpoints
    test_concurrent
    
    echo -e "\n${GREEN}✅ Load testing completed!${NC}"
    echo "Check the rate limiting behavior in the output above."
    
    # Optional recovery test (uncomment if you want to test it)
    # echo -e "\n${YELLOW}Do you want to test rate limit recovery? (30s wait) [y/N]${NC}"
    # read -r response
    # if [[ "$response" =~ ^[Yy]$ ]]; then
    #     test_recovery
    # fi
}

# Run main function
main "$@"
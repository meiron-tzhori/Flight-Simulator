#!/bin/bash

# Flight Simulator API - Curl Examples
# Usage: ./curl-examples.sh

set -e  # Exit on error

# Configuration
BASE_URL="http://localhost:8080"
COLOR_GREEN="\033[0;32m"
COLOR_BLUE="\033[0;34m"
COLOR_YELLOW="\033[1;33m"
COLOR_RED="\033[0;31m"
COLOR_RESET="\033[0m"

# Check if jq is available
if command -v jq &> /dev/null; then
    HAS_JQ=true
else
    HAS_JQ=false
    echo -e "${COLOR_YELLOW}⚠ jq not found - JSON output will not be formatted${COLOR_RESET}"
    echo -e "${COLOR_YELLOW}  Install jq: choco install jq (or download from https://stedolan.github.io/jq/)${COLOR_RESET}"
    echo ""
fi

# Helper function for JSON formatting
format_json() {
    if [ "$HAS_JQ" = true ]; then
        jq .
    else
        cat  # Just pass through without formatting
    fi
}

# Helper function for JSON field extraction
json_extract() {
    local json="$1"
    local field="$2"
    
    if [ "$HAS_JQ" = true ]; then
        echo "$json" | jq -r "$field"
    else
        # Simple fallback - just show raw JSON
        echo "$json"
    fi
}

# Helper function for printing colored output
print_header() {
    echo -e "\n${COLOR_BLUE}========================================${COLOR_RESET}"
    echo -e "${COLOR_BLUE}$1${COLOR_RESET}"
    echo -e "${COLOR_BLUE}========================================${COLOR_RESET}"
}

print_success() {
    echo -e "${COLOR_GREEN}✓ $1${COLOR_RESET}"
}

print_info() {
    echo -e "${COLOR_YELLOW}ℹ $1${COLOR_RESET}"
}

print_error() {
    echo -e "${COLOR_RED}✗ $1${COLOR_RESET}"
}

# Check if server is running
check_server() {
    print_header "Checking Server Status"
    if curl -s --max-time 2 "$BASE_URL/health" > /dev/null 2>&1; then
        print_success "Server is running at $BASE_URL"
        return 0
    else
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the simulator first: make run"
        exit 1
    fi
}

# Example 1: Health Check
example_health() {
    print_header "Example 1: Health Check"
    print_info "Command: GET /health"
    
    curl -s "$BASE_URL/health" | format_json
    
    print_success "Health check complete"
}

# Example 2: Get Initial State
example_get_state() {
    print_header "Example 2: Get Current Aircraft State"
    print_info "Command: GET /state"
    
    curl -s "$BASE_URL/state" | format_json
    
    print_success "State retrieved"
}

# Example 3: Simple Go-To Command
example_goto_simple() {
    print_header "Example 3: Go-To Command (Simple)"
    print_info "Command: POST /command/goto"
    print_info "Target: Tel Aviv (32.0853°N, 34.7818°E, 1000m)"
    
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{
            "lat": 32.0853,
            "lon": 34.7818,
            "alt": 1000.0
        }' | format_json
    
    print_success "Go-to command sent"
}

# Example 4: Go-To with Custom Speed
example_goto_speed() {
    print_header "Example 4: Go-To Command (With Speed)"
    print_info "Command: POST /command/goto"
    print_info "Target: Jerusalem (31.7683°N, 35.2137°E, 800m) at 150 m/s"
    
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{
            "lat": 31.7683,
            "lon": 35.2137,
            "alt": 800.0,
            "speed": 150.0
        }' | format_json
    
    print_success "Go-to command with speed sent"
}

# Example 5: Triangle Trajectory
example_trajectory_triangle() {
    print_header "Example 5: Trajectory Command (Triangle Pattern)"
    print_info "Command: POST /command/trajectory"
    print_info "Pattern: Triangular flight path with 3 waypoints"
    
    curl -s -X POST "$BASE_URL/command/trajectory" \
        -H "Content-Type: application/json" \
        -d '{
            "waypoints": [
                {"lat": 32.0853, "lon": 34.7818, "alt": 1500.0, "speed": 100.0},
                {"lat": 32.1053, "lon": 34.7818, "alt": 1500.0, "speed": 120.0},
                {"lat": 32.0953, "lon": 34.8018, "alt": 1500.0, "speed": 100.0}
            ]
        }' | format_json
    
    print_success "Triangle trajectory sent"
}

# Example 6: Looping Trajectory
example_trajectory_loop() {
    print_header "Example 6: Looping Trajectory (Square Pattern)"
    print_info "Command: POST /command/trajectory"
    print_info "Pattern: Square with looping enabled"
    
    curl -s -X POST "$BASE_URL/command/trajectory" \
        -H "Content-Type: application/json" \
        -d '{
            "waypoints": [
                {"lat": 32.0, "lon": 34.7, "alt": 1000.0},
                {"lat": 32.1, "lon": 34.7, "alt": 1000.0},
                {"lat": 32.1, "lon": 34.8, "alt": 1000.0},
                {"lat": 32.0, "lon": 34.8, "alt": 1000.0},
                {"lat": 32.0, "lon": 34.7, "alt": 1000.0}
            ],
            "loop": true
        }' | format_json
    
    print_success "Looping trajectory sent"
}

# Example 7: Stop Command
example_stop() {
    print_header "Example 7: Stop Command"
    print_info "Command: POST /command/stop"
    print_info "Action: Emergency stop at current position"
    
    curl -s -X POST "$BASE_URL/command/stop" | format_json
    
    print_success "Stop command sent"
}

# Example 8: Hold Command
example_hold() {
    print_header "Example 8: Hold Command"
    print_info "Command: POST /command/hold"
    print_info "Action: Orbit at current position"
    
    curl -s -X POST "$BASE_URL/command/hold" | format_json
    
    print_success "Hold command sent"
}

# Example 9: Invalid Command (Error Handling)
example_invalid() {
    print_header "Example 9: Invalid Command (Error Handling)"
    print_info "Command: POST /command/goto with invalid latitude"
    
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{
            "lat": 120.0,
            "lon": 34.7818,
            "alt": 1000.0
        }' | format_json
    
    print_info "Expected: 400 Bad Request with error details"
}

# Example 10: Monitor State Changes
example_monitor() {
    print_header "Example 10: Monitor State Changes"
    print_info "Sending command and monitoring state for 5 seconds..."
    
    # Send goto command
    print_info "Sending go-to command..."
    RESPONSE=$(curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{
            "lat": 32.1,
            "lon": 34.8,
            "alt": 1200.0,
            "speed": 100.0
        }')
    
    if [ "$HAS_JQ" = true ]; then
        echo "$RESPONSE" | jq -r '.message'
    else
        echo "$RESPONSE"
    fi
    
    # Monitor state 5 times (1 second apart)
    for i in {1..5}; do
        echo ""
        print_info "State at T+${i}s:"
        STATE=$(curl -s "$BASE_URL/state")
        
        if [ "$HAS_JQ" = true ]; then
            echo "  Position: $(echo $STATE | jq -r '.position.latitude')°N, $(echo $STATE | jq -r '.position.longitude')°E, $(echo $STATE | jq -r '.position.altitude')m"
            echo "  Speed: $(echo $STATE | jq -r '.velocity.ground_speed')m/s"
            echo "  Heading: $(echo $STATE | jq -r '.heading')°"
        else
            echo "  $STATE"
        fi
        sleep 1
    done
    
    print_success "Monitoring complete"
}

# Example 11: SSE Streaming (15 seconds)
example_stream() {
    print_header "Example 11: SSE State Streaming"
    print_info "Command: GET /stream"
    print_info "Streaming for 15 seconds (Ctrl+C to stop earlier)..."
    
    timeout 15s curl -N -s "$BASE_URL/stream" 2>/dev/null | while read -r line; do
        if [[ $line == data:* ]]; then
            # Extract JSON
            json="${line:6}"
            
            if [ "$HAS_JQ" = true ]; then
                lat=$(echo "$json" | jq -r '.position.latitude')
                lon=$(echo "$json" | jq -r '.position.longitude')
                alt=$(echo "$json" | jq -r '.position.altitude')
                speed=$(echo "$json" | jq -r '.velocity.ground_speed')
                heading=$(echo "$json" | jq -r '.heading')
                echo "[$lat°N, $lon°E, ${alt}m] Speed: ${speed}m/s, Heading: ${heading}°"
            else
                echo "$json"
            fi
        fi
    done
    
    print_success "Streaming complete"
}

# Example 12: Complete Flight Scenario
example_complete_flight() {
    print_header "Example 12: Complete Flight Scenario"
    
    print_info "Phase 1: Takeoff and climb"
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{"lat": 32.0853, "lon": 34.7818, "alt": 1000.0, "speed": 50.0}' > /dev/null
    echo "  ✓ Command sent"
    sleep 3
    
    print_info "Phase 2: Cruise to waypoint"
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{"lat": 32.1053, "lon": 34.8018, "alt": 1500.0, "speed": 120.0}' > /dev/null
    echo "  ✓ Command sent"
    sleep 3
    
    print_info "Phase 3: Hold pattern"
    curl -s -X POST "$BASE_URL/command/hold" > /dev/null
    echo "  ✓ Command sent"
    sleep 2
    
    print_info "Phase 4: Approach and descend"
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{"lat": 32.0853, "lon": 34.7818, "alt": 300.0, "speed": 60.0}' > /dev/null
    echo "  ✓ Command sent"
    sleep 3
    
    print_info "Phase 5: Final approach"
    curl -s -X POST "$BASE_URL/command/goto" \
        -H "Content-Type: application/json" \
        -d '{"lat": 32.0853, "lon": 34.7818, "alt": 50.0, "speed": 30.0}' > /dev/null
    echo "  ✓ Command sent"
    sleep 2
    
    print_info "Final state:"
    curl -s "$BASE_URL/state" | format_json
    
    print_success "Complete flight scenario finished"
}

# Main menu
show_menu() {
    clear
    echo -e "${COLOR_BLUE}"
    echo "╔══════════════════════════════════════════════════════╗"
    echo "║     Flight Simulator API - Curl Examples            ║"
    echo "╚══════════════════════════════════════════════════════╝"
    echo -e "${COLOR_RESET}"
    echo ""
    echo "Basic Examples:"
    echo "  1)  Health Check"
    echo "  2)  Get Aircraft State"
    echo "  3)  Go-To Command (Simple)"
    echo "  4)  Go-To Command (With Speed)"
    echo ""
    echo "Trajectory Examples:"
    echo "  5)  Triangle Trajectory"
    echo "  6)  Looping Trajectory (Square)"
    echo ""
    echo "Control Examples:"
    echo "  7)  Stop Command"
    echo "  8)  Hold Command"
    echo ""
    echo "Advanced Examples:"
    echo "  9)  Invalid Command (Error Demo)"
    echo "  10) Monitor State Changes"
    echo "  11) SSE State Streaming"
    echo "  12) Complete Flight Scenario"
    echo ""
    echo "Options:"
    echo "  a)  Run All Examples"
    echo "  q)  Quit"
    echo ""
    read -p "Select an example (1-12, a, q): " choice
}

# Run all examples
run_all() {
    check_server
    example_health
    sleep 1
    example_get_state
    sleep 1
    example_goto_simple
    sleep 2
    example_goto_speed
    sleep 2
    example_trajectory_triangle
    sleep 3
    example_stop
    sleep 1
    example_hold
    sleep 2
    example_invalid
    sleep 1
    
    print_header "All Examples Complete!"
    print_success "Check the output above for results"
}

# Main execution
if [ "$1" == "--all" ] || [ "$1" == "-a" ]; then
    run_all
    exit 0
fi

if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    echo "Flight Simulator API - Curl Examples"
    echo ""
    echo "Usage:"
    echo "  ./curl-examples.sh          Interactive menu"
    echo "  ./curl-examples.sh --all    Run all examples"
    echo "  ./curl-examples.sh --help   Show this help"
    echo ""
    exit 0
fi

# Interactive mode
while true; do
    show_menu
    
    case $choice in
        1)
            check_server
            example_health
            read -p "Press Enter to continue..."
            ;;
        2)
            check_server
            example_get_state
            read -p "Press Enter to continue..."
            ;;
        3)
            check_server
            example_goto_simple
            read -p "Press Enter to continue..."
            ;;
        4)
            check_server
            example_goto_speed
            read -p "Press Enter to continue..."
            ;;
        5)
            check_server
            example_trajectory_triangle
            read -p "Press Enter to continue..."
            ;;
        6)
            check_server
            example_trajectory_loop
            read -p "Press Enter to continue..."
            ;;
        7)
            check_server
            example_stop
            read -p "Press Enter to continue..."
            ;;
        8)
            check_server
            example_hold
            read -p "Press Enter to continue..."
            ;;
        9)
            check_server
            example_invalid
            read -p "Press Enter to continue..."
            ;;
        10)
            check_server
            example_monitor
            read -p "Press Enter to continue..."
            ;;
        11)
            check_server
            example_stream
            read -p "Press Enter to continue..."
            ;;
        12)
            check_server
            example_complete_flight
            read -p "Press Enter to continue..."
            ;;
        a|A)
            run_all
            read -p "Press Enter to continue..."
            ;;
        q|Q)
            echo ""
            print_info "Goodbye!"
            exit 0
            ;;
        *)
            print_error "Invalid choice. Please try again."
            sleep 1
            ;;
    esac
done

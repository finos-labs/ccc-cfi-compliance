#!/bin/bash

# Script to generate testssl.sh example outputs for all supported options
# Usage: ./generate-examples.sh <hostname>:<port>
#   e.g., ./generate-examples.sh robmoff.at:443
#   e.g., ./generate-examples.sh example.com:8080

set -e

# Check if hostname:port is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <hostname>:<port>"
    echo "Example: $0 robmoff.at:443"
    exit 1
fi

TARGET="$1"

# Extract hostname and port
if [[ $TARGET =~ ^([^:]+):([0-9]+)$ ]]; then
    HOSTNAME="${BASH_REMATCH[1]}"
    PORT="${BASH_REMATCH[2]}"
else
    echo "Error: Invalid format. Use hostname:port (e.g., robmoff.at:443)"
    exit 1
fi

# Path to testssl.sh
TESTSSL="/Users/rob/Documents/finos/ccc-general/testssl.sh/testssl.sh"

# Check if testssl.sh exists
if [ ! -f "$TESTSSL" ]; then
    echo "Error: testssl.sh not found at $TESTSSL"
    exit 1
fi

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "=================================================="
echo "Generating testssl.sh examples for $HOSTNAME:$PORT"
echo "=================================================="
echo ""

# Array of test types and their corresponding flags
# Based on Cucumber-Cloud-Language.md documentation
declare -a tests=(
    "each-cipher:--each-cipher"
    "cipher-per-proto:--cipher-per-proto"
    "std:--std"
    "forward-secrecy:-f"
    "protocols:-p"
    "grease:--grease"
    "server-defaults:-S"
    "server-preference:--server-preference"
    "vulnerable:-U"
)

# Run each test
for test_info in "${tests[@]}"; do
    # Split by colon
    IFS=':' read -r name flag <<< "$test_info"
    
    output_file="${HOSTNAME}_${PORT}_${name}.json"
    
    echo "Running: $name (flag: $flag)"
    echo "Output: $output_file"
    
    # Run testssl.sh with the appropriate flag
    if bash "$TESTSSL" "$flag" --jsonfile-pretty="$output_file" "${HOSTNAME}:${PORT}" > /dev/null 2>&1; then
        echo "✓ Generated $output_file"
    else
        echo "✗ Failed to generate $output_file (exit code: $?)"
    fi
    
    echo ""
done

echo "=================================================="
echo "Done! Generated examples in: $SCRIPT_DIR"
echo "=================================================="
echo ""
echo "Files created:"
ls -lh *.json 2>/dev/null || echo "No JSON files found"


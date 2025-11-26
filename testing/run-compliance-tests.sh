#!/bin/bash
set -euo pipefail

# CCC CFI Compliance Test Runner
# This script discovers cloud resources and runs compliance tests against them

# Default values
PROVIDER=""
OUTPUT_DIR="output"
TIMEOUT="30m"
RESOURCE_FILTER=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--provider)
      PROVIDER="$2"
      shift 2
      ;;
    -o|--output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    -t|--timeout)
      TIMEOUT="$2"
      shift 2
      ;;
    -r|--resource)
      RESOURCE_FILTER="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -p, --provider PROVIDER    Cloud provider (aws, azure, or gcp) [REQUIRED]"
      echo "  -o, --output DIR          Output directory for test reports (default: output)"
      echo "  -r, --resource RESOURCE   Filter to a specific resource name"
      echo "  -t, --timeout DURATION    Timeout for all tests (default: 30m)"
      echo "  -h, --help                Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0 --provider aws"
      echo "  $0 --provider azure --output results"
      echo "  $0 --provider aws --resource simple-inviting-kite"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

# Validate required parameters
if [ -z "$PROVIDER" ]; then
  echo "Error: --provider is required"
  echo "Use -h or --help for usage information"
  exit 1
fi

if [ "$PROVIDER" != "aws" ] && [ "$PROVIDER" != "azure" ] && [ "$PROVIDER" != "gcp" ]; then
  echo "Error: provider must be 'aws', 'azure', or 'gcp'"
  exit 1
fi

# Build the binary if needed
echo "🔨 Building compliance test runner..."
cd "$(dirname "$0")"
go build -o ccc-compliance ./runner/main.go

if [ $? -ne 0 ]; then
  echo "❌ Build failed"
  exit 1
fi

echo "✅ Build successful"
echo ""

# Build the command
CMD="./ccc-compliance -provider=\"$PROVIDER\" -output=\"$OUTPUT_DIR\" -timeout=\"$TIMEOUT\""

# Add optional flags only if set
if [ -n "$RESOURCE_FILTER" ]; then
  CMD="$CMD -resource=\"$RESOURCE_FILTER\""
fi

# Execute the command
echo "🚀 Running compliance tests..."
eval $CMD

exit $?


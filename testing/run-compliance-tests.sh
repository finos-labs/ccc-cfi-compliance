#!/bin/bash
set -euo pipefail

# CCC CFI Compliance Test Runner

# Default values
INSTANCE=""
ENV_FILE=""
SERVICE=""
OUTPUT_DIR=""
TIMEOUT="30m"
RESOURCE_FILTER=""
TAGS=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -i|--instance)
      INSTANCE="$2"
      shift 2
      ;;
    -e|--env-file)
      ENV_FILE="$2"
      shift 2
      ;;
    -s|--service)
      SERVICE="$2"
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
    -g|--tags)
      TAGS="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Required Options:"
      echo "  -i, --instance INSTANCE_ID           Instance ID from environment.yaml (e.g. main-aws, main-azure)"
      echo ""
      echo "Optional Options:"
      echo "  -e, --env-file PATH                  Path to environment.yaml (default: testing/environment.yaml)"
      echo "  -s, --service SERVICE                Service type to test. If not specified, tests all services in the instance."
      echo "                                       Valid values: object-storage, block-storage, relational-database,"
      echo "                                                     iam, load-balancer, security-group, vpc, logging"
      echo "  -o, --output DIR                     Output directory (default: testing/output)"
      echo "  -r, --resource RESOURCE              Filter to specific resource name"
      echo "  -g, --tags 'TAG1 TAG2 ...'           Space-separated tags ANDed with service tags (e.g., '@CCC.Core.CN01 @Policy')."
      echo "                                       By default @NEGATIVE and @OPT_IN scenarios are excluded."
      echo "                                       Pass '--tags @OPT_IN' to run opt-in scenarios explicitly."
      echo "  -t, --timeout DURATION               Timeout for all tests (default: 30m)"
      echo "  -h, --help                           Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0 --instance main-aws"
      echo "  $0 --instance main-azure --service object-storage"
      echo "  $0 --instance main-gcp --tags '@CCC.Core.CN04 @Policy'"
      echo "  $0 --instance main-aws --tags '@OPT_IN'               # run opt-in scenarios explicitly"
      echo "  $0 --instance main-aws --env-file /path/to/custom-environment.yaml"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

# Validate required arguments
if [ -z "$INSTANCE" ]; then
  echo "Error: --instance is required (e.g. main-aws, main-azure, main-gcp)"
  echo "Use -h or --help for usage information"
  exit 1
fi

# Build the binary
echo "🔨 Building compliance test runner..."
SCRIPT_DIR="$(dirname "$0")"
cd "$SCRIPT_DIR"
go build -o ccc-compliance ./runner/

if [ $? -ne 0 ]; then
  echo "❌ Build failed"
  exit 1
fi

echo "✅ Build successful"
echo ""

# Build the command
CMD="./ccc-compliance -instance=\"$INSTANCE\" -timeout=\"$TIMEOUT\""

if [ -n "$ENV_FILE" ]; then
  CMD="$CMD -env-file=\"$ENV_FILE\""
fi

if [ -n "$SERVICE" ]; then
  CMD="$CMD -service=\"$SERVICE\""
fi

if [ -n "$OUTPUT_DIR" ]; then
  CMD="$CMD -output=\"$OUTPUT_DIR\""
fi

if [ -n "$RESOURCE_FILTER" ]; then
  CMD="$CMD -resource=\"$RESOURCE_FILTER\""
fi

if [ -n "$TAGS" ]; then
  CMD="$CMD -tags=\"$TAGS\""
fi

# Execute the command
echo "🚀 Running compliance tests..."
eval $CMD

exit $?

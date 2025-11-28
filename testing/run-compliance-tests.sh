#!/bin/bash
set -euo pipefail

# CCC CFI Compliance Test Runner
# This script discovers cloud resources and runs compliance tests against them

# Default values
PROVIDER=""
OUTPUT_DIR=""
TIMEOUT="30m"
RESOURCE_FILTER=""
REGION=""
AZURE_SUBSCRIPTION_ID_FLAG=""
AZURE_RESOURCE_GROUP_FLAG=""
GCP_PROJECT_ID_FLAG=""

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
    --region)
      REGION="$2"
      shift 2
      ;;
    --azure-subscription-id)
      AZURE_SUBSCRIPTION_ID_FLAG="$2"
      shift 2
      ;;
    --azure-resource-group)
      AZURE_RESOURCE_GROUP_FLAG="$2"
      shift 2
      ;;
    --gcp-project-id)
      GCP_PROJECT_ID_FLAG="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Required Options:"
      echo "  -p, --provider PROVIDER              Cloud provider (aws, azure, or gcp)"
      echo ""
      echo "Optional Options:"
      echo "  -o, --output DIR                     Output directory (default: testing/output)"
      echo "  -r, --resource RESOURCE              Filter to specific resource name"
      echo "  -t, --timeout DURATION               Timeout for all tests (default: 30m)"
      echo "  --region REGION                      Cloud region"
      echo ""
      echo "Azure-specific Options (required for Azure):"
      echo "  --azure-subscription-id ID           Azure subscription ID"
      echo "  --azure-resource-group RG            Azure resource group"
      echo ""
      echo "GCP-specific Options (required for GCP):"
      echo "  --gcp-project-id PROJECT             GCP project ID"
      echo ""
      echo "  -h, --help                           Show this help message"
      echo ""
      echo "Note: All flags can also be provided via environment variables:"
      echo "  --region                 → AWS_REGION, AZURE_LOCATION, or GCP_REGION"
      echo "  --azure-subscription-id  → AZURE_SUBSCRIPTION_ID"
      echo "  --azure-resource-group   → AZURE_RESOURCE_GROUP"
      echo "  --gcp-project-id         → GCP_PROJECT_ID"
      echo ""
      echo "Examples:"
      echo "  $0 --provider aws --region us-east-1"
      echo "  $0 --provider azure --azure-subscription-id xxx --azure-resource-group cfi_test --region eastus"
      echo "  $0 --provider gcp --gcp-project-id my-project --region us-central1"
      echo ""
      echo "  # Using environment variables:"
      echo "  export AZURE_SUBSCRIPTION_ID=xxx"
      echo "  export AZURE_RESOURCE_GROUP=cfi_test"
      echo "  $0 --provider azure"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

# Fall back to environment variables if flags not provided
[ -z "$AZURE_SUBSCRIPTION_ID_FLAG" ] && AZURE_SUBSCRIPTION_ID_FLAG="${AZURE_SUBSCRIPTION_ID:-}"
[ -z "$AZURE_RESOURCE_GROUP_FLAG" ] && AZURE_RESOURCE_GROUP_FLAG="${AZURE_RESOURCE_GROUP:-}"
[ -z "$GCP_PROJECT_ID_FLAG" ] && GCP_PROJECT_ID_FLAG="${GCP_PROJECT_ID:-}"

# Set region from flag or environment based on provider
if [ -z "$REGION" ]; then
  case "$PROVIDER" in
    aws)
      REGION="${AWS_REGION:-}"
      ;;
    azure)
      REGION="${AZURE_LOCATION:-}"
      ;;
    gcp)
      REGION="${GCP_REGION:-}"
      ;;
  esac
fi

# Build the binary if needed
echo "🔨 Building compliance test runner..."
cd "$(dirname "$0")"
go build -o ccc-compliance ./runner/

if [ $? -ne 0 ]; then
  echo "❌ Build failed"
  exit 1
fi

echo "✅ Build successful"
echo ""

# Build the command
CMD="./ccc-compliance -provider=\"$PROVIDER\" -timeout=\"$TIMEOUT\""

# Add optional flags only if set
if [ -n "$OUTPUT_DIR" ]; then
  CMD="$CMD -output=\"$OUTPUT_DIR\""
fi

if [ -n "$RESOURCE_FILTER" ]; then
  CMD="$CMD -resource=\"$RESOURCE_FILTER\""
fi

if [ -n "$REGION" ]; then
  CMD="$CMD -region=\"$REGION\""
fi

if [ -n "$AZURE_SUBSCRIPTION_ID_FLAG" ]; then
  CMD="$CMD -azure-subscription-id=\"$AZURE_SUBSCRIPTION_ID_FLAG\""
fi

if [ -n "$AZURE_RESOURCE_GROUP_FLAG" ]; then
  CMD="$CMD -azure-resource-group=\"$AZURE_RESOURCE_GROUP_FLAG\""
fi

if [ -n "$GCP_PROJECT_ID_FLAG" ]; then
  CMD="$CMD -gcp-project-id=\"$GCP_PROJECT_ID_FLAG\""
fi

# Execute the command
echo "🚀 Running compliance tests..."
eval $CMD

exit $?


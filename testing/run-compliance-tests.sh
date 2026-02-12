#!/bin/bash
set -euo pipefail

# CCC CFI Compliance Test Runner
# This script discovers cloud resources and runs compliance tests against them

# Load environment variables from compliance-testing.env if it exists
SCRIPT_DIR="$(dirname "$0")"
ENV_FILE="$SCRIPT_DIR/compliance-testing.env"
if [ -f "$ENV_FILE" ]; then
  echo "📋 Loading environment from $ENV_FILE"
  set -a  # automatically export all variables
  source "$ENV_FILE"
  set +a
fi

# Default values
PROVIDER=""
SERVICE=""
OUTPUT_DIR=""
TIMEOUT="30m"
RESOURCE_FILTER=""
TAG=""
REGION=""
AZURE_SUBSCRIPTION_ID_FLAG=""
AZURE_RESOURCE_GROUP_FLAG=""
AZURE_STORAGE_ACCOUNT_FLAG=""
GCP_PROJECT_ID_FLAG=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--provider)
      PROVIDER="$2"
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
    -g|--tag)
      TAG="$2"
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
    --azure-storage-account)
      AZURE_STORAGE_ACCOUNT_FLAG="$2"
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
      echo "  -s, --service SERVICE                Service type to test. If not specified, tests all services."
      echo "                                       Valid values: object-storage, block-storage, relational-database,"
      echo "                                                     iam, load-balancer, security-group, vpc"
      echo "  -o, --output DIR                     Output directory (default: testing/output)"
      echo "  -r, --resource RESOURCE              Filter to specific resource name"
      echo "  -g, --tag TAG                        Additional tag filter ANDed with service tags (e.g., '@Policy')"
      echo "  -t, --timeout DURATION               Timeout for all tests (default: 30m)"
      echo "  --region REGION                      Cloud region"
      echo ""
      echo "Azure-specific Options (required for Azure):"
      echo "  --azure-subscription-id ID           Azure subscription ID"
      echo "  --azure-resource-group RG            Azure resource group"
      echo "  --azure-storage-account NAME         Azure storage account name"
      echo ""
      echo "GCP-specific Options (required for GCP):"
      echo "  --gcp-project-id PROJECT             GCP project ID"
      echo ""
      echo "  -h, --help                           Show this help message"
      echo ""
      echo "Note: Environment variables are auto-loaded from compliance-testing.env"
      echo "  --region                  → TF_VAR_location (Azure), TF_VAR_gcp_region (GCP), AWS_REGION"
      echo "  --azure-subscription-id   → TF_VAR_subscription_id"
      echo "  --azure-resource-group    → TF_VAR_resource_group_name"
      echo "  --azure-storage-account   → TF_VAR_storage_account_name"
      echo "  --gcp-project-id          → TF_VAR_gcp_project_id"
      echo ""
      echo "Examples:"
      echo "  $0 --provider aws --region us-east-1"
      echo "  $0 --provider aws --service object-storage --region us-east-1"
      echo "  $0 --provider azure    # uses values from compliance-testing.env"
      echo "  $0 --provider gcp --gcp-project-id my-project"
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
[ -z "$AZURE_SUBSCRIPTION_ID_FLAG" ] && AZURE_SUBSCRIPTION_ID_FLAG="${TF_VAR_subscription_id:-}"
[ -z "$AZURE_RESOURCE_GROUP_FLAG" ] && AZURE_RESOURCE_GROUP_FLAG="${TF_VAR_resource_group_name:-}"
[ -z "$AZURE_STORAGE_ACCOUNT_FLAG" ] && AZURE_STORAGE_ACCOUNT_FLAG="${TF_VAR_storage_account_name:-}"
[ -z "$GCP_PROJECT_ID_FLAG" ] && GCP_PROJECT_ID_FLAG="${TF_VAR_gcp_project_id:-}"

# Set region from flag or environment based on provider
if [ -z "$REGION" ]; then
  case "$PROVIDER" in
    aws)
      REGION="${AWS_REGION:-}"
      ;;
    azure)
      REGION="${TF_VAR_location:-}"
      ;;
    gcp)
      REGION="${TF_VAR_gcp_region:-}"
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
if [ -n "$SERVICE" ]; then
  CMD="$CMD -service=\"$SERVICE\""
fi

if [ -n "$OUTPUT_DIR" ]; then
  CMD="$CMD -output=\"$OUTPUT_DIR\""
fi

if [ -n "$RESOURCE_FILTER" ]; then
  CMD="$CMD -resource=\"$RESOURCE_FILTER\""
fi

if [ -n "$TAG" ]; then
  CMD="$CMD -tag=\"$TAG\""
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

if [ -n "$AZURE_STORAGE_ACCOUNT_FLAG" ]; then
  CMD="$CMD -azure-storage-account=\"$AZURE_STORAGE_ACCOUNT_FLAG\""
fi

if [ -n "$GCP_PROJECT_ID_FLAG" ]; then
  CMD="$CMD -gcp-project-id=\"$GCP_PROJECT_ID_FLAG\""
fi

# Execute the command
echo "🚀 Running compliance tests..."
eval $CMD

exit $?


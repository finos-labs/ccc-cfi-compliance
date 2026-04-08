#!/usr/bin/env bash
# Child-first teardown for CFI Terraform Azure stacks (remote/azure/storageaccount):
# immutability policy → blobs → container → storage account → resource group.
# Locked immutability cannot be removed via API; retention must expire or use a test config with locked=false.
set -euo pipefail

SUBSCRIPTION_ID="${1:?subscription id}"
PREFIX="${2:?resource group name prefix}"

case "$PREFIX" in
  cfi_test_*) ;;
  *)
    echo "Refusing: prefix must start with cfi_test_ (got: $PREFIX)"
    exit 1
    ;;
esac

# Fill array from a multiline string (no `mapfile < <(…)` — avoids SIGPIPE + pipefail on runners).
# Requires bash 4.3+ (`local -n`); matches `ubuntu-latest` / GitHub Actions.
read_lines_into_array() {
  local _content=$1
  local -n _target=$2
  _target=()
  while IFS= read -r _line || [[ -n "${_line}" ]]; do
    [[ -z "${_line}" ]] && continue
    _target+=("${_line}")
  done <<< "${_content}"
}

remove_container_immutability_if_possible() {
  local rg=$1 sa=$2 key=$3 container=$4
  local etag
  etag="$(az storage container immutability-policy show \
    --account-name "$sa" --container-name "$container" --account-key "$key" \
    --subscription "$SUBSCRIPTION_ID" --query etag -o tsv 2>/dev/null || true)"
  if [[ -z "$etag" || "$etag" == "null" ]]; then
    etag="$(az storage container immutability-policy show \
      --account-name "$sa" --container-name "$container" --account-key "$key" \
      --subscription "$SUBSCRIPTION_ID" --query eTag -o tsv 2>/dev/null || true)"
  fi
  if [[ -z "$etag" || "$etag" == "null" ]]; then
    return 0
  fi
  if az storage container immutability-policy delete \
    --account-name "$sa" --container-name "$container" --account-key "$key" \
    --if-match "$etag" --subscription "$SUBSCRIPTION_ID" 2>/dev/null; then
    echo "      Removed immutability policy on container $container"
    return 0
  fi
  echo "      WARN: immutability-policy delete failed (policy may be locked). Trying management-plane DELETE..."
  if az rest --method DELETE \
    --url "https://management.azure.com/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${rg}/providers/Microsoft.Storage/storageAccounts/${sa}/blobServices/default/containers/${container}/immutabilityPolicies/default?api-version=2023-05-01" \
    --headers "If-Match=${etag}" 2>/dev/null; then
    echo "      Removed immutability policy via ARM"
  else
    echo "      WARN: locked or protected immutability may block cleanup until retention expires."
  fi
}

empty_and_delete_container() {
  local sa=$1 key=$2 container=$3
  az storage blob delete-batch \
    --account-name "$sa" --account-key "$key" --subscription "$SUBSCRIPTION_ID" \
    --source "$container" --pattern '*' --delete-snapshots include 2>/dev/null || true
  az storage container delete \
    --account-name "$sa" --container-name "$container" --account-key "$key" \
    --subscription "$SUBSCRIPTION_ID" --yes 2>/dev/null || true
}

cleanup_storage_account() {
  local rg=$1 sa=$2
  local key
  key="$(az storage account keys list -g "$rg" -n "$sa" --subscription "$SUBSCRIPTION_ID" --query "[0].value" -o tsv)"
  local containers_raw containers
  containers_raw="$(az storage container list \
    --account-name "$sa" --account-key "$key" --subscription "$SUBSCRIPTION_ID" \
    --query "[].name" -o tsv 2>/dev/null || true)"
  read_lines_into_array "${containers_raw}" containers

  for c in "${containers[@]}"; do
    [[ -z "$c" ]] && continue
    echo "    Container: $c"
    remove_container_immutability_if_possible "$rg" "$sa" "$key" "$c"
    empty_and_delete_container "$sa" "$key" "$c"
  done

  echo "  Deleting storage account $sa..."
  az storage account delete --name "$sa" --resource-group "$rg" \
    --subscription "$SUBSCRIPTION_ID" --yes
}

cleanup_resource_group() {
  local rg=$1
  echo "Resource group: $rg"

  local sas_raw sas
  sas_raw="$(az storage account list -g "$rg" --subscription "$SUBSCRIPTION_ID" \
    --query "[].name" -o tsv 2>/dev/null || true)"
  read_lines_into_array "${sas_raw}" sas

  for sa in "${sas[@]}"; do
    [[ -z "$sa" ]] && continue
    echo "  Storage account: $sa"
    cleanup_storage_account "$rg" "$sa"
  done

  echo "Deleting resource group $rg (any remaining resources)..."
  az group delete --name "$rg" --subscription "$SUBSCRIPTION_ID" --yes
}

GROUPS=()
_raw_groups="$(az group list --subscription "$SUBSCRIPTION_ID" \
  --query "[?starts_with(name, '${PREFIX}')].name" -o tsv 2>/dev/null || true)"
if [[ -n "${_raw_groups//[$'\t\r\n']}" ]]; then
  _sorted_groups="$(printf '%s\n' "${_raw_groups}" | LC_ALL=C sort -u)"
  read_lines_into_array "${_sorted_groups}" GROUPS
fi

if [[ ${#GROUPS[@]} -eq 0 || -z "${GROUPS[0]:-}" ]]; then
  echo "No resource groups match prefix '$PREFIX'."
  exit 0
fi

echo "Matching resource groups (${#GROUPS[@]}):"
printf '  - %s\n' "${GROUPS[@]}"
echo ""

for rg in "${GROUPS[@]}"; do
  [[ -z "$rg" ]] && continue
  cleanup_resource_group "$rg"
  echo ""
done

echo "Cleanup finished."

#!/usr/bin/env bash
# Child-first teardown for CFI Terraform Azure stacks (remote/azure/storageaccount).
#
# Azure locked + version-level immutability: blob versions must exit their retention window
# (see testing/environment.yaml object-storage-retention-period-days and Terraform
# object_storage_retention_period_days) before the immutability policy can be removed and the
# storage account deleted. This script is intended to run on a schedule (e.g. daily): failures
# to delete an account or RG are treated as deferrals — the next run often succeeds once
# retention and blob soft-delete periods have passed.
#
# Per-container order: delete blobs (best effort) → remove immutability policy → delete blobs
# again → delete container → storage account → resource group.
set -uo pipefail

SUBSCRIPTION_ID="${1:?subscription id}"
PREFIX="${2:?resource group name prefix}"

case "$PREFIX" in
  cfi_test_*) ;;
  *)
    echo "Refusing: prefix must start with cfi_test_ (got: $PREFIX)"
    exit 1
    ;;
esac

# Set to 1 if any storage account or RG delete is deferred (run again later).
CLEANUP_DEFERRED=0

# Fill array from a multiline string (no `mapfile < <(…)` — avoids SIGPIPE + pipefail on CI).
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
    --if-match "$etag" --subscription "$SUBSCRIPTION_ID"; then
    echo "      Removed immutability policy on container $container"
    return 0
  fi
  echo "      WARN: data-plane immutability-policy delete failed; trying ARM DELETE..."
  if az rest --method DELETE \
    --url "https://management.azure.com/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${rg}/providers/Microsoft.Storage/storageAccounts/${sa}/blobServices/default/containers/${container}/immutabilityPolicies/default?api-version=2023-05-01" \
    --headers "If-Match=${etag}"; then
    echo "      Removed immutability policy via ARM"
    return 0
  fi
  echo "      WARN: Immutability policy still present (blobs in retention or locked policy)." >&2
  echo "      Scheduled cleanup will retry; see object-storage-retention-period-days in environment.yaml." >&2
  return 0
}

delete_container_blobs() {
  local sa=$1 key=$2 container=$3
  az storage blob delete-batch \
    --account-name "$sa" --account-key "$key" --subscription "$SUBSCRIPTION_ID" \
    --source "$container" --pattern '*' --delete-snapshots include 2>/dev/null || true
}

delete_storage_container() {
  local sa=$1 key=$2 container=$3
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
    # Locked immutability: clear blobs first (no-op until retention ends), then try policy removal.
    delete_container_blobs "$sa" "$key" "$c"
    remove_container_immutability_if_possible "$rg" "$sa" "$key" "$c"
    delete_container_blobs "$sa" "$key" "$c"
    delete_storage_container "$sa" "$key" "$c"
  done

  echo "  Deleting storage account $sa..."
  if az storage account delete --name "$sa" --resource-group "$rg" \
    --subscription "$SUBSCRIPTION_ID" --yes; then
    return 0
  fi
  echo "  WARN: Storage account $sa not deleted yet (immutability/versioning or dependencies). Will retry on next run." >&2
  CLEANUP_DEFERRED=1
  return 1
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
    cleanup_storage_account "$rg" "$sa" || true
  done

  # Without --no-wait, az blocks until every resource is gone. A protected storage account can
  # leave that wait running for a very long time. --no-wait returns once the delete is queued;
  # Azure keeps working in the background; the next scheduled run retries if the RG still exists.
  echo "Deleting resource group $rg (any remaining resources, async)..."
  if az group delete --name "$rg" --subscription "$SUBSCRIPTION_ID" --yes --no-wait; then
    echo "  Resource group delete started (async). If protected resources remain, the RG may persist until a later run."
    return 0
  fi
  echo "WARN: Could not start delete for resource group $rg. Will retry on next run." >&2
  CLEANUP_DEFERRED=1
  return 1
}

# Do not use array name GROUPS — bash sets $GROUPS to numeric supplementary group IDs (read-only-ish);
# reusing it produced bogus "resource group" names like 1001, 100, 118 in CI.
CFI_RG_NAMES=()
_raw_groups="$(az group list --subscription "$SUBSCRIPTION_ID" \
  --query "[?starts_with(name, '${PREFIX}')].name" -o tsv 2>/dev/null || true)"
if [[ -n "${_raw_groups//[$'\t\r\n']}" ]]; then
  _sorted_groups="$(printf '%s\n' "${_raw_groups}" | LC_ALL=C sort -u)"
  read_lines_into_array "${_sorted_groups}" CFI_RG_NAMES
fi

if [[ ${#CFI_RG_NAMES[@]} -eq 0 || -z "${CFI_RG_NAMES[0]:-}" ]]; then
  echo "No resource groups match prefix '$PREFIX'."
  exit 0
fi

echo "Matching resource groups (${#CFI_RG_NAMES[@]}):"
printf '  - %s\n' "${CFI_RG_NAMES[@]}"
echo ""

for rg in "${CFI_RG_NAMES[@]}"; do
  [[ -z "$rg" ]] && continue
  cleanup_resource_group "$rg" || true
  echo ""
done

echo "Cleanup finished."
if [[ "${CLEANUP_DEFERRED}" -ne 0 ]]; then
  echo "NOTE: Some deletes were deferred (immutability retention / soft-delete). Daily scheduled runs should complete teardown after those windows pass." >&2
fi
exit 0

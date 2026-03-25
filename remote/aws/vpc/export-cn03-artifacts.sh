#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="${1:-.}"
MATRIX_FILE="${OUT_DIR%/}/cn03-peer-trials.json"
ENV_FILE="${OUT_DIR%/}/cn03-feature.env"

mkdir -p "${OUT_DIR}"

terraform output -json cn03_peer_trial_matrix > "${MATRIX_FILE}"

python3 - "${ENV_FILE}" <<'PY'
import json
import shlex
import subprocess
import sys

env_output = subprocess.check_output(
    ["terraform", "output", "-json", "cn03_feature_env"],
    text=True,
)
env_map = json.loads(env_output)

with open(sys.argv[1], "w", encoding="utf-8") as env_file:
    env_file.write("# CN03 feature env exports (safe to source in interactive shells)\n")
    for key in sorted(env_map):
        env_file.write(f"export {key}={shlex.quote(str(env_map[key]))}\n")
PY

MATRIX_ABS="$(cd "$(dirname "${MATRIX_FILE}")" && pwd)/$(basename "${MATRIX_FILE}")"
echo "export CN03_PEER_TRIAL_MATRIX_FILE='${MATRIX_ABS}'" >> "${ENV_FILE}"

echo "Wrote ${MATRIX_FILE}"
echo "Wrote ${ENV_FILE}"

#!/bin/bash
set -e
set -o pipefail

# Resolve directory containing this script (works when invoked via relative path)
SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
# shellcheck source="./common.sh"
source "$SCRIPT_DIR/common.sh"

set_home
run_hook_dir "init"
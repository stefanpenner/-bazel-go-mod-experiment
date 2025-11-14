#!/bin/bash
set -euo pipefail

MODULE_STAGING="${GO_MOD_DIR}"

if [ ! -d "$MODULE_STAGING" ]; then
  echo "Error: GO_MOD_DIR path is not a directory: $MODULE_STAGING"
  exit 1
fi

shopt -s nullglob
MODULE_CANDIDATES=("$MODULE_STAGING"/github.com/example/multipkg@*)
shopt -u nullglob

if [ "${#MODULE_CANDIDATES[@]}" -eq 0 ]; then
  echo "Error: Module directory not found under $MODULE_STAGING"
  exit 1
fi

MODULE_DIR="${MODULE_CANDIDATES[0]}"

check_file() {
  local file="$1"
  if [ ! -f "$MODULE_DIR/$file" ]; then
    echo "Error: expected file $file not found in staged module"
    exit 1
  fi
}

check_file "go.mod"
check_file "main.go"
check_file "alpha/alpha.go"
check_file "beta/beta.go"

if ! grep -q "module github.com/example/multipkg" "$MODULE_DIR/go.mod"; then
  echo "Error: go.mod does not contain expected module path"
  cat "$MODULE_DIR/go.mod"
  exit 1
fi

echo "multi_pkg staged module looks good"

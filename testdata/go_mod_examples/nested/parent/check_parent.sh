#!/bin/bash
set -euo pipefail

MODULE_STAGING="${GO_MOD_DIR}"

if [ ! -d "$MODULE_STAGING" ]; then
  echo "Error: GO_MOD_DIR path is not a directory: $MODULE_STAGING"
  exit 1
fi

shopt -s nullglob
MODULE_CANDIDATES=("$MODULE_STAGING"/github.com/example/nested/parent@*)
shopt -u nullglob

if [ "${#MODULE_CANDIDATES[@]}" -eq 0 ]; then
  echo "Error: Module directory not found under $MODULE_STAGING"
  exit 1
fi

MODULE_DIR="${MODULE_CANDIDATES[0]}"

if [ ! -f "$MODULE_DIR/parent.go" ]; then
  echo "Error: parent.go not found in staged parent module"
  exit 1
fi

if [ -e "$MODULE_DIR/child/child.go" ]; then
  echo "Error: child module sources leaked into parent staging directory"
  exit 1
fi

echo "nested parent staging looks good"

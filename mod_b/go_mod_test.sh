#!/bin/bash
set -euo pipefail

MODULE_STAGING="${GO_MOD_DIR:-}"

if [ ! -d "$MODULE_STAGING" ]; then
  echo "Error: GO_MOD_DIR path is not a directory: $MODULE_STAGING"
  exit 1
fi

shopt -s nullglob
MODULE_CANDIDATES=("$MODULE_STAGING"/mod_b@*)
shopt -u nullglob

if [ "${#MODULE_CANDIDATES[@]}" -eq 0 ]; then
  echo "Error: Module directory not found under $MODULE_STAGING"
  exit 1
fi

MODULE_DIR="${MODULE_CANDIDATES[0]}"

if [ ! -d "$MODULE_DIR" ]; then
  echo "Error: Expected module directory missing: $MODULE_DIR"
  exit 1
fi

GO_MOD_FILE="$MODULE_DIR/go.mod"
if [ ! -f "$GO_MOD_FILE" ]; then
  echo "Error: go.mod not found in $MODULE_DIR"
  exit 1
fi

if [ ! -f "$MODULE_DIR/lib.go" ]; then
  echo "Error: $MODULE_DIR/lib.go not found, but was expected"
  exit 1
fi

if ! grep -q "module github.com/stefanpenner/-bazel-go-mod-experiment/mod_b" "$GO_MOD_FILE"; then
  echo "Error: go.mod has incorrect module path"
  cat "$GO_MOD_FILE"
  exit 1
fi

if ! grep -q "package mod_b" "$MODULE_DIR/lib.go"; then
  echo "Error: $MODULE_DIR/lib.go has incorrect package declaration"
  cat "$MODULE_DIR/lib.go"
  exit 1
fi

echo "All tests passed!"


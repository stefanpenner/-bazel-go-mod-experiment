#!/bin/bash

# Get the path to the zip file
ZIP_FILE="$GO_MOD"

# Create a temporary directory for extraction
TMP_DIR="$TEST_TMPDIR/mod_b_test"

# Extract the zip file
unzip -q "$ZIP_FILE" -d "$TMP_DIR"


# Verify the module structure
MODULE_DIR=($TMP_DIR/github.com/stefanpenner/-bazel-go-mod-experiment/mod_b@*)

if [ -d "$file" ]; then
    echo "Error: Module directory not found: $MODULE_DIR"
    exit 1
fi

if [ ! -d "$MODULE_DIR" ]; then
    echo "Error: Module directory not found: $MODULE_DIR"
    exit 1
fi

# Verify go.mod exists and has correct content
GO_MOD="$MODULE_DIR/go.mod"
if [ ! -f "$GO_MOD" ]; then
    echo "Error: go.mod not found"
    exit 1
fi

# Verify source files exist
if [ ! -f "$MODULE_DIR/lib.go" ]; then
    echo "Error: $MODULE_DIR/lib.go not found, but was expected"
    exit 1
fi

# Verify go.mod content
if ! grep -q "module github.com/stefanpenner/-bazel-go-mod-experiment/mod_b" "$GO_MOD"; then
    echo "Error: go.mod has incorrect module path"
    cat "$MODULE_DIR/$GO_MOD"
    exit 1
fi

# Verify main.go content
if ! grep -q "package mod_b" "$MODULE_DIR/lib.go"; then
    echo "Error: $MODULE_DIR/lib.go has incorrect package declaration"
    cat "$MODULE_DIR/lib.go"
    exit 1
fi

echo "All tests passed!" 
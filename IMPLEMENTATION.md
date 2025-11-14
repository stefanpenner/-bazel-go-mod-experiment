# go_mod Rule Implementation

This document describes the implementation of the `go_mod` rule that generates Go module directories from `go_library` targets.

## Overview

The `go_mod` rule takes all `go_library` targets within a Go module and produces a Bazel directory containing:
1. The `go.mod` file
2. All source files from the `go_library` targets, preserving their directory structure

### Key Features

1. **go_library as deps**: The rule accepts `go_library` targets as dependencies and automatically extracts their source files
2. **Directory output**: Produces a directory with loose files instead of a zip archive
3. **Source-only dependencies**: Only depends on source files (`.go` files), ensuring the rule becomes dirty only when source files change
4. **Gazelle integration**: Automatically discovers `go_library` targets and generates `go_mod` rules

## Implementation Details

### 1. Rule Definition (`rules/go_mod.bzl`)

The `go_mod` rule has been updated with the following changes:

#### Attributes
- `go_mod`: Label pointing to the `go.mod` file (required)
- `deps`: List of `go_library` targets to include (optional)
- `srcs`: Additional source files or filegroups (optional, for backwards compatibility)
- `module_path`: The Go module path (required)

#### Implementation
1. **Source Collection**: Extracts source files from `go_library` targets using `GoInfo.sources`
2. **Directory Creation**: Uses `ctx.actions.declare_directory()` to create an output directory
3. **File Copying**: Generates a shell script that copies files to the output directory, preserving structure
4. **Path Handling**: Strips package prefixes to get relative paths within the module

### 2. Gazelle Extension (`gazelle_languages/go_mod/extension.go`)

The gazelle extension has been updated to automatically discover `go_library` targets:

#### Approach
1. **State Tracking**: Maintains a list of all discovered `go_library` targets during traversal
2. **Collection Phase**: For each package, collects all `go_library` rules from the BUILD file
3. **Association Phase**: When a `go.mod` is found, associates all `go_library` targets whose package path starts with the module path
4. **Generation**: Generates a `go_mod` rule with the collected `go_library` targets as `deps`

#### Key Methods
- `GenerateRules()`: Discovers `go_library` targets and generates `go_mod` rules
- Leverages gazelle's depth-first post-order traversal to collect targets before generating rules

## Usage

### Manual Usage

```starlark
load("@bazel-go-mod-experiment//rules:go_mod.bzl", "go_mod")

go_mod(
    name = "go_mod_zip",
    deps = [
        ":my_lib",
        "//pkg/subpkg:sublib",
    ],
    go_mod = ":go.mod",
    module_path = "github.com/example/mymodule",
)
```

### With Gazelle

1. Ensure your `go_library` targets exist in BUILD files
2. Run gazelle to generate `go_mod` rules:

```bash
bazel run //:gazelle
```

Gazelle will:
- Find all packages with a `go.mod` file
- Discover all `go_library` targets in that module
- Generate or update the `go_mod` rule with the correct `deps`

## Benefits

### 1. Source-Only Dependencies
The rule only depends on source files, not compiled artifacts. This means:
- Changes to source files trigger rebuilds
- Changes to compiled libraries (e.g., test files, generated files) don't trigger unnecessary rebuilds
- Faster incremental builds

### 2. Directory Output
Producing a directory instead of a zip:
- Easier to inspect and debug
- Can be directly consumed by other rules
- Simpler to integrate with CI/CD pipelines

### 3. Automatic Discovery
Gazelle automatically discovers `go_library` targets, reducing manual maintenance:
- No need to manually list all packages in the module
- Automatically updates when new packages are added
- Consistent with other Gazelle workflows

## Testing

The implementation includes testdata with two example modules:

### simple_module
- Basic module with one subpackage
- Demonstrates simple usage

### complex_module
- Module with multiple subpackages
- Includes internal packages
- Demonstrates nested package structures

## Future Improvements

1. **Module Root Detection**: Currently assumes each `go.mod` defines a separate module. Could be enhanced to handle nested modules.
2. **Dependency Tracking**: Could track external dependencies and include them in the output.
3. **Version Management**: Could integrate with version manifest for publishing.
4. **Go Sum**: Could include `go.sum` file handling.

## Migration Guide

### From Old Implementation

The old implementation used filegroups as `srcs`. To migrate:

**Before:**
```starlark
go_mod(
    name = "go_mod_zip",
    srcs = [
        ":_pkg_",
        "//pkg/subpkg:_pkg_",
    ],
    go_mod = ":go.mod",
    module_path = "mod_a",
)
```

**After:**
```starlark
go_mod(
    name = "go_mod_zip",
    deps = [
        ":my_lib",
        "//pkg/subpkg:sublib",
    ],
    go_mod = ":go.mod",
    module_path = "mod_a",
)
```

Or simply run `bazel run //:gazelle` to automatically generate the new format.

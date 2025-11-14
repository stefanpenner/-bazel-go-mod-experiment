# Testdata for go_mod Rule

This directory contains test examples for the `go_mod` rule that generates Go module directories from `go_library` targets.

## Directory Structure

### simple_module
A basic Go module with:
- A main package at the root
- One subpackage (`pkg`)
- Two `go_library` targets

### complex_module
A more complex Go module with:
- A main package at the root
- Multiple subpackages (`api`, `internal/util`)
- Three `go_library` targets
- Internal packages with restricted visibility

## How the go_mod Rule Works

The `go_mod` rule takes all `go_library` targets within a Go module and produces a directory containing:
1. The `go.mod` file
2. All source files from the `go_library` targets, preserving directory structure

### Example Usage

```starlark
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

### Key Features

1. **Source-only dependencies**: The rule only depends on source files (`.go` files), not compiled artifacts
2. **Directory output**: Produces a directory with loose files instead of a zip archive
3. **Incremental builds**: Only rebuilds when source files change
4. **Gazelle integration**: Use gazelle to automatically discover `go_library` targets and generate `go_mod` rules

## Running Gazelle

To automatically generate `go_mod` rules from existing `go_library` targets:

```bash
bazel run //:gazelle
```

This will:
1. Discover all `go_library` targets in packages with a `go.mod` file
2. Generate or update `go_mod` rules with the correct `deps`
3. Set the `module_path` based on the directory path

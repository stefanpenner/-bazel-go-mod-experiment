load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    strip_prefix = ctx.attr.strip_prefix

    # Collect source files from go_library targets
    # Only include .go source files, not non-source files like BUILD.bazel, README.md, etc.
    go_source_files = []
    all_inputs = []
    
    for src in ctx.attr.srcs:
        # If it's a go_library target, get the GoInfo
        if GoInfo in src:
            go_info = src[GoInfo]
            # Get source files from the go_library (only direct sources, not transitive)
            source_files = go_info.source.srcs.to_list()
            go_source_files.extend(source_files)
            # Track inputs for dependency tracking (so rule becomes dirty when sources change)
            all_inputs.append(go_info.source.srcs)
            # Also track transitive go files for dependency tracking, but don't include them in output
            all_inputs.append(go_info.transitive_go_files)
        elif DefaultInfo in src:
            # If it's a filegroup or file, filter to only .go files
            files = src[DefaultInfo].files.to_list()
            for f in files:
                if f.extension == "go":
                    go_source_files.append(f)
                    all_inputs.append(depset([f]))

    if not go_source_files:
        fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    # Output directory instead of zip
    output_dir = ctx.actions.declare_directory(ctx.attr.name)

    # Collect all inputs: go.mod, stamp file (if any), and all source files
    inputs = [go_mod]
    stamp = maybe_stamp(ctx)
    if stamp:
        inputs.append(stamp.volatile_status_file)
    all_inputs_depset = depset(inputs, transitive=all_inputs)

    go_mod_tool = ctx.executable._go_mod_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", ctx.label.package)
    args.add("--output", output_dir.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    if stamp:
        args.add("--volatile-status-file", stamp.volatile_status_file.path)

    # Pass only .go source files
    for src in go_source_files:
        args.add("--src", src.path)

    ctx.actions.run(
        outputs=[output_dir],
        inputs=all_inputs_depset,
        executable=go_mod_tool,
        arguments=[args],
        progress_message="Creating Go module directory %s" % ctx.label,
    )

    return [DefaultInfo(files=depset([output_dir]))]

_go_mod = rule(
  implementation = _go_mod_impl,
  attrs = dict({
    "go_mod": attr.label(
      mandatory = True,
      allow_single_file = True,
      doc = "The go.mod file for the module",
    ),
    "strip_prefix": attr.string(
      doc = "Prefix to strip from source file paths",
    ),
    "srcs": attr.label_list(
      providers = [[GoInfo], [DefaultInfo]],
      doc = "go_library targets to include in the module. Only source files from these targets will be included.",
    ),
    "module_path": attr.string(
      mandatory = True,
      doc = "The module path (e.g., github.com/my_project)",
    ),
    "_go_mod_tool": attr.label(
      default = "//go_mod_tool:go_mod_tool",
      executable = True,
      cfg = "exec",
      doc = "Go executable to create the module directory",
    ),
  }, **STAMP_ATTRS),
  doc = "Creates a Go module directory containing loose files for the bundled go.mod. Only depends on source files, not non-source files.",
)

def go_mod(name, go_mod, srcs, module_path, visibility = None):
  _go_mod(
    name = name,
    go_mod = go_mod,
    srcs = srcs,
    module_path = module_path,
    visibility = visibility
  )

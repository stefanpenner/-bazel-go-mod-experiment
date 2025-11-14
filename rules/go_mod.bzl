load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    strip_prefix = ctx.attr.strip_prefix

    # Collect source files from go_library targets (GoInfo.source_files)
    # Only include .go source files, not BUILD files or other non-source files
    source_files_depsets = []
    for src in ctx.attr.srcs:
        if GoInfo in src:
            # Extract source files from go_library target
            go_info = src[GoInfo]
            source_files_depsets.append(go_info.source_files)
        elif DefaultInfo in src:
            # Fallback for filegroups or direct file inputs
            # Filter to only .go files
            all_files = src[DefaultInfo].files.to_list()
            go_files = [f for f in all_files if f.basename.endswith(".go")]
            if go_files:
                source_files_depsets.append(depset(go_files))
    
    all_srcs = depset(transitive=source_files_depsets)

    if not all_srcs.to_list():
        fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    # Output a directory instead of a zip
    output_dir = ctx.actions.declare_directory(ctx.attr.name)

    # Collect all inputs: go.mod, stamp file (if any), and all source files
    # Note: We only include source files, not BUILD files or other non-source files
    inputs = [go_mod]
    stamp = maybe_stamp(ctx)
    if stamp:
        inputs.append(stamp.volatile_status_file)
    all_inputs = depset(inputs, transitive=[all_srcs])

    go_mod_tool = ctx.executable._go_mod_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", ctx.label.package)
    args.add("--output", output_dir.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    if stamp:
        args.add("--volatile-status-file", stamp.volatile_status_file.path)

    # Pass all source files as arguments
    for src in all_srcs.to_list():
        args.add("--src", src.path)

    ctx.actions.run(
        outputs=[output_dir],
        inputs=all_inputs,
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
      providers = [[GoInfo], []],
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
  doc = "Creates a Go module directory containing loose files for the bundled go.mod. Only depends on source files from go_library targets.",
)

def go_mod(name, go_mod, srcs, module_path, visibility = None):
  _go_mod(
    name = name,
    go_mod = go_mod,
    srcs = srcs,
    module_path = module_path,
    visibility = visibility
  )

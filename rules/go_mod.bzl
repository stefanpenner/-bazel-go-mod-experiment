load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path

    # Collect source files from go_library deps
    src_files = []
    for dep in ctx.attr.deps:
        if GoInfo in dep:
            # Extract source files from go_library targets
            go_info = dep[GoInfo]
            for src in go_info.sources:
                src_files.append(src)
    
    # Also collect files from srcs attribute (for backwards compatibility)
    for src in ctx.attr.srcs:
        for f in src[DefaultInfo].files.to_list():
            src_files.append(f)

    if not src_files:
        fail("No source files found in deps or srcs: %s" % (ctx.attr.deps + ctx.attr.srcs))

    # Create output directory
    output_dir = ctx.actions.declare_directory(ctx.attr.name + "_dir")

    # Collect all inputs
    inputs = [go_mod]
    stamp = maybe_stamp(ctx)
    if stamp:
        inputs.append(stamp.volatile_status_file)
    
    all_inputs = depset(inputs + src_files)

    # Create a script to copy files to the output directory
    script = ctx.actions.declare_file(ctx.attr.name + "_gen.sh")
    script_content = ["#!/bin/bash", "set -euo pipefail", ""]
    
    # Create the output directory
    script_content.append("mkdir -p " + output_dir.path)
    
    # Copy go.mod
    script_content.append("cp " + go_mod.path + " " + output_dir.path + "/go.mod")
    
    # Copy all source files, preserving directory structure
    for src in src_files:
        # Strip the package prefix to get relative path
        rel_path = src.path
        if src.path.startswith(ctx.label.package + "/"):
            rel_path = src.path[len(ctx.label.package) + 1:]
        
        # Create parent directory if needed
        parent_dir = rel_path.rsplit("/", 1)[0] if "/" in rel_path else ""
        if parent_dir:
            script_content.append("mkdir -p " + output_dir.path + "/" + parent_dir)
        
        script_content.append("cp " + src.path + " " + output_dir.path + "/" + rel_path)
    
    ctx.actions.write(
        output = script,
        content = "\n".join(script_content),
        is_executable = True,
    )

    ctx.actions.run(
        outputs=[output_dir],
        inputs=all_inputs,
        executable=script,
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
    "deps": attr.label_list(
      providers = [[GoInfo]],
      doc = "go_library targets to include in the module",
    ),
    "srcs": attr.label_list(
      doc = "Additional source files or filegroups to include (for backwards compatibility)",
    ),
    "module_path": attr.string(
      mandatory = True,
      doc = "The module path (e.g., github.com/my_project)",
    ),
  }, **STAMP_ATTRS),
  doc = "Creates a Go module directory with loose files for bundling",
)

def go_mod(name, go_mod, module_path, deps = None, srcs = None, visibility = None):
  _go_mod(
    name = name,
    go_mod = go_mod,
    deps = deps or [],
    srcs = srcs or [],
    module_path = module_path,
    visibility = visibility
  )

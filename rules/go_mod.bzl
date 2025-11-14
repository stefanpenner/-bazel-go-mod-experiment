load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_archive_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    strip_prefix = ctx.attr.strip_prefix or ctx.label.package

    srcs_depsets = []
    for dep in ctx.attr.deps:
        if GoInfo not in dep:
            fail("All deps must provide GoInfo. %s does not." % dep.label)
        srcs_depsets.append(dep[GoInfo].srcs)

    all_srcs = depset(transitive=srcs_depsets)
    src_files = all_srcs.to_list()
    if not src_files:
        fail("No .go source files found in deps: %s" % ctx.attr.deps)

    output_dir = ctx.actions.declare_directory(ctx.attr.name + "_files")

    inputs = [go_mod]
    stamp = maybe_stamp(ctx)
    if stamp:
        inputs.append(stamp.volatile_status_file)
    all_inputs = depset(inputs, transitive=[all_srcs])

    go_mod_tool = ctx.executable._go_mod_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", strip_prefix)
    args.add("--output-dir", output_dir.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    if stamp:
        args.add("--volatile-status-file", stamp.volatile_status_file.path)

    for src in src_files:
        args.add("--src", src.path)

    ctx.actions.run(
        outputs=[output_dir],
        inputs=all_inputs,
        executable=go_mod_tool,
        arguments=[args],
        progress_message="Staging Go module files %s" % ctx.label,
    )

    return [DefaultInfo(files=depset([output_dir]))]

_go_mod = rule(
  implementation = _go_mod_archive_impl,
  attrs = dict({
    "go_mod": attr.label(
      mandatory = True,
      allow_single_file = True,
      doc = "The go.mod file for the module",
    ),
    "strip_prefix": attr.string(
      doc = "Workspace-relative prefix to strip from staged files; defaults to the Bazel package path.",
    ),
    "deps": attr.label_list(
      mandatory = True,
      providers = [GoInfo],
      doc = "Go library targets whose source files should be staged into the module directory.",
    ),
    "module_path": attr.string(
      mandatory = True,
      doc = "The logical module path (e.g., github.com/my_project).",
    ),
    "_go_mod_tool": attr.label(
      default = "//go_mod_tool:go_mod_tool",
      executable = True,
      cfg = "exec",
      doc = "Go executable to stage the module files.",
    ),
  }, **STAMP_ATTRS),
  doc = "Stages loose Go module files for bundling by copying the go.mod and dependent Go source files into a directory.",
)

def go_mod(name, go_mod, module_path, deps, strip_prefix = None, visibility = None):
  _go_mod(
    name = name,
    go_mod = go_mod,
    module_path = module_path,
    deps = deps,
    strip_prefix = strip_prefix,
    visibility = visibility,
  )

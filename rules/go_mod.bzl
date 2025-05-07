load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_archive_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    strip_prefix = ctx.attr.strip_prefix

    # Collect all files from srcs (could be filegroups, go_library, etc.)
    srcs_depsets = [src[DefaultInfo].files for src in ctx.attr.srcs]
    all_srcs = depset(transitive=srcs_depsets)

    if not all_srcs:
        fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    output_zip = ctx.actions.declare_file(ctx.attr.name + ".zip")

    # Collect all inputs: go.mod, stamp file (if any), and all srcs
    inputs = [go_mod]
    stamp = maybe_stamp(ctx)
    if stamp:
        inputs.append(stamp.volatile_status_file)
    all_inputs = depset(inputs, transitive=[all_srcs])

    go_mod_tool = ctx.executable._go_mod_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", strip_prefix)
    args.add("--output", output_zip.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    if stamp:
        args.add("--volatile-status-file", stamp.volatile_status_file.path)

    # If you need to pass all srcs as arguments, you must convert to a list
    for src in all_srcs.to_list():
        args.add("--src", src.path)

    ctx.actions.run(
        outputs=[output_zip],
        inputs=all_inputs,
        executable=go_mod_tool,
        arguments=[args],
        progress_message="Creating Go module archive %s" % ctx.label,
    )

    return [DefaultInfo(files=depset([output_zip]))]

_go_mod = rule(
  implementation = _go_mod_archive_impl,
  attrs = dict({
    "go_mod": attr.label(
      mandatory = True,
      allow_single_file = True,
      doc = "The go.mod file for the module",
    ),
    "strip_prefix": attr.string(

    ),
    "srcs": attr.label_list(
      providers = [[GoInfo], []],
      doc = "Go source files or go_library targets to include in the module archive",
    ),
    "module_path": attr.string(
      mandatory = True,
      doc = "The module path (e.g., github.com/my_project)",
    ),
    "_go_mod_tool": attr.label(
      default = "//go_mod_tool:go_mod_tool",
      executable = True,
      cfg = "exec",
      doc = "Go executable to create the module archive",
    ),
  }, **STAMP_ATTRS),
  doc = "Creates a Go module archive (.zip) for use with a Go proxy",
)

def go_mod(name, go_mod, srcs, module_path, strip_prefix = None, visibility = None):
  _go_mod(
    name = name,
    go_mod = go_mod,
    srcs = srcs,
    module_path = module_path,
    strip_prefix = strip_prefix,
    visibility = visibility
  )

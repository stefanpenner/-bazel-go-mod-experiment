load("@rules_go//go:def.bzl", "GoInfo")
load("@aspect_bazel_lib//lib:stamping.bzl", "STAMP_ATTRS", "maybe_stamp")

def _go_mod_archive_impl(ctx):
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    strip_prefix = ctx.attr.strip_prefix
    inputs = [go_mod]

    stamp = maybe_stamp(ctx)
    # if stamping is enabled, ensure the volatile_status_file is part of the
    # tracked inputs (todo: is that needed, since it's volatile?, presumably
    # this is important for the stable one, but not the volatile one)
    if stamp:
      inputs.append(stamp.volatile_status_file)

    # Collect source files from go_library targets
    srcs = []

    print("ctx.attr.srcs", ctx.attr.srcs)
    for src in ctx.attr.srcs:
      if GoInfo in src:
          go_info = src[GoInfo]
          print("addings srcs of", src)
      else:
          print("skipping srcs of", src)
          continue

      # append direct sources
      srcs.extend(go_info.srcs)

      # append sources of depencies
      for dep in go_info.deps:
        # dep.data is GoArchiveData provider
        go_archive_data = dep.data

        print("candidate", go_archive_data.importpath)
        # include include deps who have an import path that contains the
        # current module_path
        if go_archive_data.importpath.startswith(module_path):
          srcs.extend(go_archive_data.srcs)

    if not srcs:
      fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    output_zip = ctx.actions.declare_file(ctx.attr.name + ".zip")

    inputs = inputs + srcs
    print("srcs:", srcs)
    go_mod_tool = ctx.executable._go_mod_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", strip_prefix)
    args.add("--output", output_zip.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    args.add("--volatile-status-file", stamp.volatile_status_file.path)

    for src in srcs:
      args.add("--src", src.path)
    ctx.actions.run(
      outputs = [output_zip],
      inputs = inputs,
      executable = go_mod_tool,
      arguments = [args],
      progress_message = "Creating Go module archive %s" % ctx.label,
    )

    return [DefaultInfo(files = depset([output_zip]))]

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

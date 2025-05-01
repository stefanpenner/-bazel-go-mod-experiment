load("@rules_go//go:def.bzl", "GoInfo")

def _go_mod_archive_impl(ctx):
    # Inputs
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    version = ctx.attr.version
    strip_prefix = ctx.attr.strip_prefix

    # Collect source files from go_library targets
    srcs = []
    for src in ctx.attr.srcs:
      go_info = src[GoInfo]

      # append direct sources
      srcs.extend(go_info.srcs)

      # append sources of depencies
      for dep in go_info.deps:
        # dep.data is GoArchiveData provider
        go_archive_data = dep.data

        # include include deps who have an import path that contains the
        # current module_path
        if go_archive_data.importpath.startswith(module_path):
          srcs.extend(go_archive_data.srcs)

    if not srcs:
        fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    output_zip = ctx.actions.declare_file(ctx.attr.name + ".zip")

    inputs = [go_mod] + srcs
    go_tool = ctx.executable._archive_tool

    args = ctx.actions.args()
    args.add("--strip-prefix", strip_prefix)
    args.add("--output", output_zip.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    args.add("--version", version)
    for src in srcs:
        args.add("--src", src.path)

    ctx.actions.run(
        outputs = [output_zip],
        inputs = inputs,
        executable = go_tool,
        arguments = [args],
        progress_message = "Creating Go module archive %s" % ctx.label,
    )

    return [DefaultInfo(files = depset([output_zip]))]

_go_mod = rule(
    implementation = _go_mod_archive_impl,
    attrs = {
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
        "version": attr.string(
            mandatory = True,
            doc = "The module version (e.g., v0.1.0)",
        ),
        "_archive_tool": attr.label(
            default = "//archive_tool:archive_tool",
            executable = True,
            cfg = "exec",
            doc = "Go executable to create the module archive",
        ),
    },
    doc = "Creates a Go module archive (.zip) for use with a Go proxy",
)
def go_mod(name, go_mod, srcs, module_path, strip_prefix = None, visibility = None):
    _go_mod(
        name = name,
        go_mod = go_mod,
        srcs = srcs,
        module_path = module_path,
        version = "{VOLATILE_VERSION}" if "{VOLATILE_VERSION}" else '0.0.0',
        strip_prefix = strip_prefix,
        visibility = visibility,
    )

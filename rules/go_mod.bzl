
load("@rules_go//go:def.bzl", "GoArchive", "GoInfo")

go_library = provider(
    fields = {
        "srcs": "Source files of the go_library",
        "importpath": "Import path of the go_library",
    }
)

def _go_mod_archive_impl(ctx):
    # Inputs
    go_mod = ctx.file.go_mod
    module_path = ctx.attr.module_path
    version = ctx.attr.version

    # Collect source files from go_library targets
    srcs = []
    for src in ctx.attr.srcs:
      go_info = src[GoInfo]

      # append direct sources
      srcs.extend(go_info.srcs)

      # append sources of depencies
      for dep in go_info.deps:
        srcs.extend(dep.source.srcs)

    if not srcs:
        fail("No .go source files found in srcs: %s" % ctx.attr.srcs)

    # Output: the .zip file
    output_zip = ctx.actions.declare_file(ctx.attr.name + ".zip")

    # Inputs for the go executable
    inputs = [go_mod] + srcs
    go_tool = ctx.executable._archive_tool

    # Arguments for the go executable
    args = ctx.actions.args()
    args.add("--output", output_zip.path)
    args.add("--module-path", module_path)
    args.add("--go-mod", go_mod.path)
    args.add("--version", version)
    for src in srcs:
        args.add("--src", src.path)

    # Run the go executable
    ctx.actions.run(
        outputs = [output_zip],
        inputs = inputs,
        executable = go_tool,
        arguments = [args],
        progress_message = "Creating Go module archive %s" % ctx.label,
    )

    return [DefaultInfo(files = depset([output_zip]))]

go_mod = rule(
    implementation = _go_mod_archive_impl,
    attrs = {
        "go_mod": attr.label(
            mandatory = True,
            allow_single_file = True,
            doc = "The go.mod file for the module",
        ),
        "srcs": attr.label_list(
            providers = [[go_library], []],
            # allow_files = [".go"],
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

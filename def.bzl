load("@bazel_skylib//lib:paths.bzl", "paths")

def _prototool_impl(ctx):
    commands = []

    # Prototool works better from relative paths, so cd to the directroy where
    # the action was invoked.
    commands.append("cd \"$BUILD_WORKING_DIRECTORY\"")

    # Invoke prototool with the user arguments.
    abs_prototool_path = paths.join("\"$BUILD_WORKSPACE_DIRECTORY\"", ctx.executable._prototool.path)
    commands.append("{0} $@".format(abs_prototool_path))

    ctx.actions.run_shell(
        outputs = [ctx.outputs.executable],
        command = "echo '{commands}' > {output}".format(
            commands = " && ".join(commands),
            output = ctx.outputs.executable.path,
        ),
        arguments = ["$@"],
        tools = [ctx.executable._prototool],
    )

    return DefaultInfo(executable = ctx.outputs.executable)

prototool = rule(
    implementation = _prototool_impl,
    executable = True,
    attrs = {
        "_prototool": attr.label(
            cfg = "host",
            default = Label("//cmd/prototool"),
            executable = True,
        ),
    },
)

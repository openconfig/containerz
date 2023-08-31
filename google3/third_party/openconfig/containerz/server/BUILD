package(
    default_visibility = ["//third_party/openconfig/containerz:users"],
    licenses = ["notice"],
)

go_library(
    name = "server",
    srcs = [
        "deploy.go",
        "list.go",
        "logs.go",
        "options.go",
        "remove.go",
        "server.go",
        "start.go",
        "stop.go",
    ],
    deps = [
        "//third_party/openconfig/containerz/chunker",
        "//third_party/openconfig/containerz/containers:options",
    ],
)

go_test(
    name = "server_test",
    size = "small",
    srcs = [
        "deploy_test.go",
        "list_test.go",
        "logs_test.go",
        "options_test.go",
        "remove_test.go",
        "start_test.go",
        "stop_test.go",
    ],
    library = ":server",
    deps = ["//third_party/openconfig/containerz/containers:options"],
)

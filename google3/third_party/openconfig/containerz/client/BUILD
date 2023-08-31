package(
    default_visibility = ["//third_party/openconfig/containerz:users"],
    licenses = ["notice"],
)

go_library(
    name = "client",
    srcs = [
        "client.go",
        "client_google.go",
        "list.go",
        "logs.go",
        "pull.go",
        "push.go",
        "remove.go",
        "start.go",
        "stop.go",
        "types.go",
    ],
    deps = [
        "//base/go:google",
        "//net/grpc/go:grpcprod",
        "//third_party/openconfig/containerz/chunker",
    ],
)

go_test(
    name = "client_test",
    size = "small",
    srcs = [
        "client_test.go",
        "list_test.go",
        "logs_test.go",
        "pull_test.go",
        "push_test.go",
        "remove_test.go",
        "start_test.go",
        "stop_test.go",
        "types_test.go",
    ],
    library = ":client",
    deps = ["//net/grpc/go:grpctest"],
)

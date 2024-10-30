load("//tools/build_defs/go:go_library.bzl", "go_library")
load("//tools/build_defs/go:go_test.bzl", "go_test")

package(licenses = ["notice"])

go_library(
    name = "client",
    srcs = [
        "client.go",
        "create_volume.go",
        "list_container.go",
        "list_image.go",
        "list_volume.go",
        "logs.go",
        "pull.go",
        "push.go",
        "remove_container.go",
        "remove_image.go",
        "remove_volume.go",
        "start_container.go",
        "stop_container.go",
        "types.go",
        "update_container.go",
    ],
    deps = ["//net/sdn/decentralized/appsupport/security"],
)

go_test(
    name = "client_test",
    size = "small",
    srcs = [
        "client_test.go",
        "create_volume_test.go",
        "list_container_test.go",
        "list_image_test.go",
        "list_volume_test.go",
        "logs_test.go",
        "pull_test.go",
        "push_test.go",
        "remove_container_test.go",
        "remove_image_test.go",
        "remove_volume_test.go",
        "start_container_test.go",
        "stop_container_test.go",
        "types_test.go",
        "update_container_test.go",
    ],
    library = ":client",
)

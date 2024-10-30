load("//tools/build_defs/go:go_library.bzl", "go_library")
load("//tools/build_defs/go:go_test.bzl", "go_test")

package(licenses = ["notice"])

go_library(
    name = "server",
    srcs = [
        "create_volume.go",
        "deploy.go",
        "list_container.go",
        "list_image.go",
        "list_volume.go",
        "logs.go",
        "options.go",
        "remove_container.go",
        "remove_image.go",
        "remove_volume.go",
        "server.go",
        "start_container.go",
        "stop_container.go",
        "update_container.go",
    ],
    deps = ["//net/sdn/decentralized/appsupport/security"],
)

go_test(
    name = "server_test",
    size = "small",
    srcs = [
        "create_volume_test.go",
        "deploy_test.go",
        "list_container_test.go",
        "list_image_test.go",
        "list_volume_test.go",
        "logs_test.go",
        "options_test.go",
        "remove_container_test.go",
        "remove_image_test.go",
        "remove_volume_test.go",
        "start_container_test.go",
        "stop_container_test.go",
        "update_container_test.go",
    ],
    library = ":server",
)

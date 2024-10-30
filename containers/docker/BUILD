load("//tools/build_defs/go:go_library.bzl", "go_library")
load("//tools/build_defs/go:go_test.bzl", "go_test")

package(licenses = ["notice"])

go_library(
    name = "docker",
    srcs = [
        "container_list.go",
        "container_logs.go",
        "container_remove.go",
        "container_start.go",
        "container_stop.go",
        "container_update.go",
        "image_list.go",
        "image_pull.go",
        "image_push.go",
        "image_remove.go",
        "janitor.go",
        "manager.go",
        "volume_create.go",
        "volume_list.go",
        "volume_remove.go",
    ],
)

go_test(
    name = "docker_test",
    size = "small",
    srcs = [
        "container_list_test.go",
        "container_logs_test.go",
        "container_remove_test.go",
        "container_start_test.go",
        "container_stop_test.go",
        "container_update_test.go",
        "image_list_test.go",
        "image_pull_test.go",
        "image_push_test.go",
        "image_remove_test.go",
        "janitor_test.go",
        "manager_test.go",
        "volume_create_test.go",
        "volume_list_test.go",
        "volume_remove_test.go",
    ],
    library = ":docker",
)

package(licenses = ["notice"])

go_library(
    name = "client",
    srcs = [
        "client.go",
        "create_volume.go",
        "list_container.go",
        "list_volume.go",
        "logs.go",
        "pull.go",
        "push.go",
        "remove_container.go",
        "remove_volume.go",
        "start_container.go",
        "stop_container.go",
        "types.go",
    ],
)

go_test(
    name = "client_test",
    size = "small",
    srcs = [
        "client_test.go",
        "create_volume_test.go",
        "list_container_test.go",
        "list_volume_test.go",
        "logs_test.go",
        "pull_test.go",
        "push_test.go",
        "remove_container_test.go",
        "remove_volume_test.go",
        "start_container_test.go",
        "stop_container_test.go",
        "types_test.go",
    ],
    library = ":client",
)

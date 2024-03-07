package(licenses = ["notice"])

go_library(
    name = "server",
    srcs = [
        "create_volume.go",
        "deploy.go",
        "list_container.go",
        "list_volume.go",
        "logs.go",
        "options.go",
        "remove_container.go",
        "remove_volume.go",
        "server.go",
        "start_container.go",
        "stop_container.go",
    ],
)

go_test(
    name = "server_test",
    size = "small",
    srcs = [
        "create_volume_test.go",
        "deploy_test.go",
        "list_container_test.go",
        "list_volume_test.go",
        "logs_test.go",
        "options_test.go",
        "remove_container_test.go",
        "remove_volume_test.go",
        "start_container_test.go",
        "stop_container_test.go",
    ],
    library = ":server",
)

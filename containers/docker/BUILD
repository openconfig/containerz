package(licenses = ["notice"])

go_library(
    name = "docker",
    srcs = [
        "container_list.go",
        "container_logs.go",
        "container_start.go",
        "container_stop.go",
        "image_pull.go",
        "image_push.go",
        "image_remove.go",
        "janitor.go",
        "manager.go",
    ],
)

go_test(
    name = "docker_test",
    size = "small",
    srcs = [
        "container_list_test.go",
        "container_logs_test.go",
        "container_start_test.go",
        "container_stop_test.go",
        "image_pull_test.go",
        "image_push_test.go",
        "image_remove_test.go",
        "janitor_test.go",
        "manager_test.go",
    ],
    library = ":docker",
)

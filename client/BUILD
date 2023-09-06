package(licenses = ["notice"])

go_library(
    name = "client",
    srcs = [
        "client.go",
        "list.go",
        "logs.go",
        "pull.go",
        "push.go",
        "remove.go",
        "start.go",
        "stop.go",
        "types.go",
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
)

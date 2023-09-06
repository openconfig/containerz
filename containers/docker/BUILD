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
        "manager.go",
    ],
    deps = [
        "//third_party/golang/dockerconnections/nat",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/container",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/filters",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/network",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/registry",
        "//third_party/golang/github_com/moby/moby/v/v24/pkg/jsonmessage",
        "//third_party/golang/opencontainers/image_spec/specs_go/v1",
        "//third_party/openconfig/containerz/containers:options",
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
        "manager_test.go",
    ],
    library = ":docker",
    deps = [
        "//third_party/golang/dockerconnections/nat",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/container",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/filters",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/network",
        "//third_party/golang/github_com/moby/moby/v/v24/api/types/registry",
        "//third_party/golang/github_com/moby/moby/v/v24/pkg/jsonmessage",
        "//third_party/golang/opencontainers/image_spec/specs_go/v1",
        "//third_party/openconfig/containerz/containers:options",
    ],
)

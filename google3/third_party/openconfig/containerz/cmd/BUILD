package(licenses = ["notice"])

go_library(
    name = "cmd",
    srcs = [
        "container.go",
        "container_list.go",
        "container_logs.go",
        "container_start.go",
        "container_stop.go",
        "image.go",
        "image_pull.go",
        "image_push.go",
        "image_remove.go",
        "root.go",
        "start.go",
    ],
    deps = [
        "//third_party/golang/github_com/moby/moby/v/v24/client",
        "//third_party/openconfig/containerz/client",
        "//third_party/openconfig/containerz/containers/docker",
        "//third_party/openconfig/containerz/server",
    ],
)

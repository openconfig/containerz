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
        "volume.go",
        "volume_create.go",
        "volume_list.go",
        "volume_remove.go",
    ],
    deps = [
        "//base/go:google",
        "//net/grpc/go:grpcprod",
    ],
)

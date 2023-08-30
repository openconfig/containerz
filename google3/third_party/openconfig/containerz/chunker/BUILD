package(
    default_visibility = ["//third_party/openconfig/containerz:users"],
    licenses = ["notice"],
)

go_library(
    name = "chunker",
    srcs = [
        "reader.go",
        "writer.go",
    ],
)

go_test(
    name = "chunker_test",
    size = "small",
    srcs = [
        "reader_test.go",
        "writer_test.go",
    ],
    library = ":chunker",
)

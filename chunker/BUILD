load("//tools/build_defs/go:go_library.bzl", "go_library")
load("//tools/build_defs/go:go_test.bzl", "go_test")

package(licenses = ["notice"])

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

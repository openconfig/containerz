load("//tools/build_defs/go:go_library.bzl", "go_library")
load("//tools/build_defs/go:go_test.bzl", "go_test")

package(licenses = ["notice"])

go_library(
    name = "options",
    srcs = ["options.go"],
)

go_test(
    name = "options_test",
    size = "small",
    srcs = ["options_test.go"],
    library = ":options",
)

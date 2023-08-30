package(
    default_visibility = ["//third_party/openconfig/containerz:users"],
    licenses = ["notice"],
)

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

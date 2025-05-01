module github.com/stefanpenner/-bazel-go-mod-experiment/mod_a

go 1.23.3

require github.com/stefanpenner/-bazel-go-mod-experiment/mod_b v0.0.0

replace github.com/stefanpenner/-bazel-go-mod-experiment/mod_b => ../mod_b/

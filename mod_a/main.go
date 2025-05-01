package main

import (
	"fmt"

	"github.com/stefanpenner/-bazel-go-mod-experiment/mod_a/foo"
	"github.com/stefanpenner/-bazel-go-mod-experiment/mod_b"
)

func main() {
	fmt.Printf("%s, World!\n", foo.Hello())
	fmt.Printf("%s, World!\n", mod_b.Hello())
}

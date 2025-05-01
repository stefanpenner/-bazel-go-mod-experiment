package main

import (
	"fmt"

	"github.com/stefanpenner/-bazel-go-mod-experiment/mod_a/foo"
)

func main() {
	fmt.Printf("Hello, %s!", foo.Hello())
}

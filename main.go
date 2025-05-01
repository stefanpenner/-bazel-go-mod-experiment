package main

import (
	"fmt"

	"github.com/stefanpenner/-bazel-go-mod-experiment/foo"
)

func main() {
	fmt.Printf("Hello, %s!", foo.Hello())
}

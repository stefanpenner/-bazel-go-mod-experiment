package main

import (
	"fmt"
	"github.com/example/test_module_nested/pkg1"
	"github.com/example/test_module_nested/pkg2"
)

func main() {
	fmt.Println(pkg1.Foo())
	fmt.Println(pkg2.Bar())
}

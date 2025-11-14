package main

import (
	"fmt"
	"github.com/example/module1/pkg1"
	"github.com/example/module1/pkg2"
)

func main() {
	fmt.Println("Module1")
	pkg1.Hello()
	pkg2.World()
}

package main

import (
	"fmt"
	"github.com/example/complex_module/api"
	"github.com/example/complex_module/internal/util"
)

func main() {
	fmt.Println("Complex module")
	api.HandleRequest()
	util.Helper()
}

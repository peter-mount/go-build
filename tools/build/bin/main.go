package main

import (
	"fmt"
	_ "github.com/peter-mount/go-build/tools/build"
	"github.com/peter-mount/go-kernel/v2"
	"os"
)

func main() {
	if err := kernel.Launch(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package main

import (
	"os"

	"github.com/oa-plugins/webauto/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

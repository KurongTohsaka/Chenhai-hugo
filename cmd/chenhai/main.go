package main

import (
	"os"

	"github.com/KurongTohsaka/chenhai-hugo/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

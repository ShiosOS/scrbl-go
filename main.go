package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/juliuswalton/scrbl/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

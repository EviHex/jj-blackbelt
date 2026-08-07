package main

import (
	"context"
	"fmt"
	"os"

	"github.com/EviHex/jj-blackbelt/internal/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "bb: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pinglei-he/blackbelt/internal/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "blackbelt: %v\n", err)
		os.Exit(1)
	}
}

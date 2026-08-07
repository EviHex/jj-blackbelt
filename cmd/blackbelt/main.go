package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/pinglei-he/blackbelt/internal/blackbelt"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "print one terminal diagram without changing GitHub comments")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "blackbelt: no positional arguments are supported")
		os.Exit(2)
	}
	if err := blackbelt.Run(context.Background(), blackbelt.Options{DryRun: *dryRun}); err != nil {
		fmt.Fprintf(os.Stderr, "blackbelt: %v\n", err)
		os.Exit(1)
	}
}

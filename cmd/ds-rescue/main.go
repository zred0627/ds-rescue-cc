package main

import (
	"flag"
	"fmt"
	"os"
)

var Version = "0.0.1-dev"

func main() {
	version := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("ds-rescue v%s\n", Version)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "ds-rescue: not yet implemented (skeleton)")
	os.Exit(1)
}

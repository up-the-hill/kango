package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/up-the-hill/kango/internal/parser"
	"github.com/up-the-hill/kango/internal/ui"
)

func main() {
	// args
	file := flag.String("f", "", "kanban.md file path")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "Error: no file path provided")
		os.Exit(1)
	}

	f, err := parser.ParseFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ui.Main(f, *file)
}

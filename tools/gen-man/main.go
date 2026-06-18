package main

import (
	"log"
	"os"

	"github.com/jamesjohnsdev/issues/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "man"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	header := &doc.GenManHeader{
		Title:   "ISSUES",
		Section: "1",
		Source:  "issues",
	}
	if err := doc.GenManTree(cmd.Root(), header, dir); err != nil {
		log.Fatal(err)
	}
}

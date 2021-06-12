package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse the command line arguments.
	if len(os.Args) > 2 {
		fmt.Println("usage: igor [directory]")
		os.Exit(1)
	}
	inDir := "."
	if len(os.Args) > 1 {
		inDir = os.Args[1]
	}

	// Walk the directory and transpile all the Igor files.
	err := filepath.Walk(inDir, func(inFile string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !strings.HasSuffix(inFile, ".igor") {
			return nil
		}
		outFile := inFile[:len(inFile)-5] + ".go"
		return transpile(inFile, outFile)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

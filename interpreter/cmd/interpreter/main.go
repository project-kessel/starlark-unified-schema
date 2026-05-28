package main

import (
	"flag"
	"fmt"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
)

func main() {
	srcDirArg := flag.String("src", "schema", "The path to the directory containing Starlark schema (.star) files. (Default: schema)")
	flag.Parse()

	loader := lang.NewLoader(*srcDirArg)
	processor := lang.NewProcessor(loader)

	for _, file_path := range flag.Args() {
		err := processor.ProcessModule(file_path, nil)
		if err != nil {
			fmt.Printf("Error processing module %s: %s", file_path, err)
		}
	}
}

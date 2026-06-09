package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/project-kessel/starlark-unified-schema/internal/jsonschema"
	"github.com/project-kessel/starlark-unified-schema/internal/lang"
)

func main() {
	srcDir := flag.String("src", "schema", "path to the directory containing Starlark schema (.star) files")
	outputDir := flag.String("output-dir", "output", "directory to write generated artifacts")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Processes Starlark schema files and generates JSON Schema artifacts.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	loader := lang.NewLoader(*srcDir)
	processor := lang.NewProcessor(loader)

	resources, err := processor.ProcessAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing resources: %v\n", err)
		os.Exit(1)
	}

	if len(resources) == 0 {
		fmt.Println("No schemas generated.")
		return
	}

	visitor := jsonschema.NewVisitor()
	for i := range resources {
		resources[i].Accept(visitor)
	}

	fmt.Printf("Writing schemas to %s/\n", *outputDir)
	if err := jsonschema.WriteSchemas(*outputDir, visitor.Outputs); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schemas: %v\n", err)
		os.Exit(1)
	}
}

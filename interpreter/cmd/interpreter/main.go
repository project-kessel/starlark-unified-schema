package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
)

func main() {
	srcDir := flag.String("src", "schema", "path to the directory containing Starlark schema (.star) files")
	outputDir := flag.String("output-dir", "output", "directory to write generated artifacts")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file ...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Processes Starlark schema files and generates JSON Schema artifacts.\n")
		fmt.Fprintf(os.Stderr, "If no files are specified, all .star files under -src are processed.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	loader := lang.NewLoader(*srcDir)
	processor := lang.NewProcessor(loader)

	jsonSchemaVisitor := output.NewJSONSchemaVisitor()
	if err := processor.Process(jsonSchemaVisitor, flag.Args()...); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing schema: %v\n", err)
		os.Exit(1)
	}

	results := jsonSchemaVisitor.Results()
	if len(results) == 0 {
		fmt.Println("No schemas generated.")
		return
	}

	fmt.Printf("Writing schemas to %s/\n", *outputDir)
	if err := output.WriteSchemas(*outputDir, results); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schemas: %v\n", err)
		os.Exit(1)
	}
}

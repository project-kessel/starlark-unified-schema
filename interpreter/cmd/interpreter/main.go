package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
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

	registry := lang.NewResourceRegistry()
	loader := lang.NewLoader(*srcDir, registry)
	thread := &starlark.Thread{
		Name:  "main",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
		Load:  loader.Load,
	}

	moduleNames, err := loader.GetModuleNames()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving module names: %v\n", err)
		os.Exit(1)
	}

	for _, name := range moduleNames {
		_, err := loader.Load(thread, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading module %s: %v\n", name, err)
			os.Exit(1)
		}
	}

	jsonSchemaVisitor := output.NewJSONSchemaVisitor()
	if err := lang.VisitResources(registry.Resources(), jsonSchemaVisitor); err != nil {
		fmt.Fprintf(os.Stderr, "Error visiting resources: %v\n", err)
		os.Exit(1)
	}

	if len(jsonSchemaVisitor.Outputs) == 0 {
		fmt.Println("No schemas generated.")
		return
	}

	fmt.Printf("Writing schemas to %s/\n", *outputDir)
	if err := output.WriteSchemas(*outputDir, jsonSchemaVisitor.Outputs); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schemas: %v\n", err)
		os.Exit(1)
	}
}

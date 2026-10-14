package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"github.com/project-kessel/starlark-unified-schema/internal/output/ksil"
)

func main() {
	// Dispatch happens before any flag parsing so that bare invocation and
	// `interpreter [-src dir] [file ...]` keep their existing meaning.
	if isDocsInvocation(os.Args) {
		if err := runDocs(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "docs: %v\n", err)
			os.Exit(1)
		}
		return
	}

	runCompile()
}

func runCompile() {
	srcDir := flag.String("src", "schema", "path to the directory containing Starlark schema (.star) files")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [file ...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s %s [options]\n\n", os.Args[0], docsSubcommand)
		fmt.Fprintf(os.Stderr, "Processes Starlark schema files and generates JSON Schema artifacts.\n")
		fmt.Fprintf(os.Stderr, "If no files are specified, all .star files under -src are processed.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s\tGenerate a Markdown API reference from the DSL docstrings.\n\n", docsSubcommand)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	loader := lang.NewLoader(*srcDir)
	processor := lang.NewProcessor(loader)

	directoryToVisitorMappings := map[string]output.SchemaVisitor{
		"JSONSCHEMA_OUTPUT_DIR": output.NewJSONSchemaVisitor(),
		"KSL_OUTPUT_DIR":        ksil.NewKSILVisitor(),
	}
	outputConfigs := createOutputConfigs(directoryToVisitorMappings)
	if len(outputConfigs) == 0 {
		keys := make([]string, 0, len(directoryToVisitorMappings))
		for key := range directoryToVisitorMappings {
			keys = append(keys, key)
		}
		fmt.Fprintln(os.Stderr, "No output configured. Set one or more of the following environment variables:", strings.Join(keys, ", "))
		os.Exit(1)
	}

	inputFiles := flag.Args()

	for _, config := range outputConfigs {
		if err := processVisitorAndWriteOutputs(processor, config, inputFiles); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing visitor and writing outputs: %v\n", err)
			os.Exit(1)
		}
	}
}

func processVisitorAndWriteOutputs(processor *lang.Processor, config OutputConfig, files []string) error {
	if err := processor.Process(config.Visitor, files...); err != nil {
		return fmt.Errorf("error processing visitor: %w", err)
	}

	results, err := config.Visitor.Results()
	if err != nil {
		return fmt.Errorf("error getting results from visitor: %w", err)
	}

	if err := output.WriteSchemas(config.Path, results); err != nil {
		return fmt.Errorf("error writing schemas to %s: %w", config.Path, err)
	}

	return nil
}

func createOutputConfigs(mappings map[string]output.SchemaVisitor) []OutputConfig {
	configs := []OutputConfig{}

	for varName, visitor := range mappings {
		path := os.Getenv(varName)
		if path == "" {
			continue
		}
		configs = append(configs, OutputConfig{
			Path:    path,
			Visitor: visitor,
		})
	}

	return configs
}

type OutputConfig struct {
	Path    string
	Visitor output.SchemaVisitor
}

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

	outputConfigs := createOutputConfigs()
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

func createOutputConfigs() []OutputConfig {
	configs := []OutputConfig{}
	addConfig := func(varName string, visitor output.SchemaVisitor) {
		path := os.Getenv(varName)
		if path == "" {
			return
		}
		configs = append(configs, OutputConfig{
			Path:    path,
			Visitor: visitor,
		})
	}

	addConfig("JSONSCHEMA_OUTPUT_DIR", output.NewJSONSchemaVisitor())
	addConfig("KSL_OUTPUT_DIR", output.NewKSILVisitor())

	return configs
}

type OutputConfig struct {
	Path    string
	Visitor output.SchemaVisitor
}

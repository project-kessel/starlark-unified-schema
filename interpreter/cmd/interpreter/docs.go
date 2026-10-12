package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/project-kessel/starlark-unified-schema/internal/docs"
	"github.com/project-kessel/starlark-unified-schema/internal/lang"
)

const docsSubcommand = "docs"

// isDocsInvocation reports whether args (an os.Args-shaped slice) selects the
// docs subcommand.
//
// Only the first argument is considered, and only as an exact match. Bare
// invocation means "compile everything" and must stay that way, and every
// positional argument the compile path takes is a .star path, so a schema file
// can never be mistaken for a subcommand.
func isDocsInvocation(args []string) bool {
	return len(args) > 1 && args[1] == docsSubcommand
}

func runDocs(args []string) error {
	fs := flag.NewFlagSet(docsSubcommand, flag.ExitOnError)
	srcDir := fs.String("src", "schema", "path to the directory containing Starlark schema (.star) files")
	outPath := fs.String("o", "schema/API.md", "path to write the generated Markdown reference")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s [options]\n\n", os.Args[0], docsSubcommand)
		fmt.Fprintf(os.Stderr, "Generates a Markdown API reference from the docstrings in the Starlark DSL.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	reference, err := docs.Generate(lang.NewLoader(*srcDir))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil { //rwxr-xr-x
		return fmt.Errorf("creating output directory for %s: %w", *outPath, err)
	}

	if err := os.WriteFile(*outPath, []byte(docs.Markdown(reference)), 0o644); err != nil { //rw-r--r--
		return fmt.Errorf("writing %s: %w", *outPath, err)
	}

	return nil
}

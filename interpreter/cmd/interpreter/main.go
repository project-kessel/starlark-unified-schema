package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/project-kessel/starlark-unified-schema/internal/lang"
	"github.com/project-kessel/starlark-unified-schema/internal/output"
)

func main() {
	srcDirArg := flag.String("src", "schema", "The path to the directory containing Starlark schema (.star) files.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input-file(s)]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Reads the input files (if given) relative to the src directory and generates outputs. If no input files are given, all .star files in the src directory are processed.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  input file(s): Starlark files to process (relative to the src directory.) Leave empty to process all files.\n")
	}

	flag.Parse()

	loader := lang.NewLoader(*srcDirArg)
	processor := lang.NewProcessor(loader)

	visitors := []output.Visitor{output.NewKSILVisitor()}
	paths := []string{"outputs/ksil/"}

	src_files := flag.Args()

	for i, visitor := range visitors {
		if len(src_files) > 0 {
			for _, src_file := range src_files {
				err := processor.ProcessModule(src_file, visitor)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		} else {
			err := processor.ProcessAllModules(visitor)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		outputs, err := visitor.GetOutput()
		if err != nil {
			fmt.Println(err)
			return
		}

		path := paths[i]
		output.WriteOutputs(path, outputs)
	}
}

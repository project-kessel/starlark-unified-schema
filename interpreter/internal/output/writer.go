package output

import (
	"fmt"
	"os"
	"path/filepath"
)

type OutputEntry struct {
	Path     string
	Contents []byte
}

func WriteSchemas(outputDir string, entries []OutputEntry) error {
	for _, entry := range entries {
		fullPath := filepath.Join(outputDir, entry.Path)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, entry.Contents, 0644); err != nil {
			return fmt.Errorf("error writing %s: %w", fullPath, err)
		}

		fmt.Printf("  wrote %s\n", entry.Path)
	}

	return nil
}

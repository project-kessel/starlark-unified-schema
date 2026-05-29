package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteSchemas(outputDir string, entries []OutputEntry) error {
	for _, entry := range entries {
		fullPath := filepath.Join(outputDir, entry.Path)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}

		data, err := json.MarshalIndent(entry.Schema, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling schema for %s: %w", entry.Path, err)
		}
		data = append(data, '\n')

		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			return fmt.Errorf("error writing %s: %w", fullPath, err)
		}

		fmt.Printf("  wrote %s\n", entry.Path)
	}

	return nil
}

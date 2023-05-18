package module

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Preload(m *ModuleLoader) error {
	moduleFiles, err := os.ReadDir(m.config.ModuleDirectory)
	if err != nil {
		return fmt.Errorf("failed to read module directory: %w", err)
	}

	preloadCount := 0

	for _, file := range moduleFiles {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".wasm" {
			moduleName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			if _, err := m.LoadModule(moduleName); err != nil {
				return fmt.Errorf("failed to load module %s: %w", moduleName, err)
			}

			preloadCount++
		}
	}

	log.Printf("preloaded %d modules.\n", preloadCount)

	return nil
}

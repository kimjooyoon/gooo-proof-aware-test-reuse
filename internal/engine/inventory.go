package engine

import (
	"os"
	"path/filepath"
	"strings"
)

func BuildInventory(root string) (Inventory, error) {
	inventory := Inventory{
		RootReadmeExcluded:   true,
		GitExcluded:          true,
		CallerOutputExcluded: true,
		CacheExcluded:        true,
		VendorExcluded:       true,
		ToolchainExcluded:    true,
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if entry.IsDir() && excludedDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			inventory.DescendantDirs++
			return nil
		}
		if relative == "README.md" || !entry.Type().IsRegular() {
			return nil
		}
		inventory.RegularFiles++
		switch filepath.Ext(entry.Name()) {
		case ".go":
			inventory.GoFiles++
		case ".gooo":
			inventory.GoooFiles++
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		inventory.PhysicalLines += physicalLineCount(data)
		return nil
	})
	return inventory, err
}

func excludedDirectory(name string) bool {
	switch strings.ToLower(name) {
	case ".git", ".cache", "cache", "vendor", "toolchain", "toolchains", "node_modules":
		return true
	default:
		return strings.Contains(strings.ToLower(name), "toolchain")
	}
}

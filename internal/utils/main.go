package utils

import "path/filepath"

func ExtensionlessBasename(path string) string {
	base := filepath.Base(path)
	name := base[:len(base)-len(filepath.Ext(base))]
	return name
}

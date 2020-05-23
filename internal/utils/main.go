package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/dropseed/workhorse/internal/meta"
)

func ExtensionlessBasename(path string) string {
	base := filepath.Base(path)
	name := base[:len(base)-len(filepath.Ext(base))]
	return name
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Find(dir, name, extension string) string {
	if fileExists(name) {
		return name
	}
	return path.Join(meta.AppName, dir, fmt.Sprintf("%s.%s", name, extension))
}

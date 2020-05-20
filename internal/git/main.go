package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func CleanBranchName(s string) string {
	branch := strings.ToLower(s)
	if !strings.HasPrefix(branch, "workhorse/") {
		branch = "workhorse/" + branch
	}
	if ext := filepath.Ext(branch); ext != "" {
		branch = branch[:len(branch)-len(ext)]
	}
	return branch
}

func Branch(to string) {
	cmd := exec.Command("git", "checkout", "-b", to)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

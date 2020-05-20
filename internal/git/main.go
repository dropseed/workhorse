package git

import (
	"os"
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
	cmd := exec.Command("git", "checkout", "-B", to)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func IsDirty() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	cleaned := strings.TrimSpace(string(out))
	return cleaned != ""
}

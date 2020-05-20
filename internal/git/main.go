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

func Checkout(name string) error {
	cmd := exec.Command("git", "checkout", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func CreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func DeleteBranch(name string) error {
	cmd := exec.Command("git", "branch", "-D", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func Status() string {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return string(out)
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

func Commit(path, message string) {
	cmd := exec.Command("git", "add", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func Push(branch string) {
	cmd := exec.Command("git", "push", "--force", "--set-upstream", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func Remote() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	remote, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	s := string(remote)
	s = strings.TrimSpace(s)
	return s
}

func LastCommitFilesAdded(filterPrefix string) []string {
	cmd := exec.Command("git", "diff", "HEAD^", "HEAD", "--name-only", "--diff-filter", "A")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	s := string(out)

	lines := strings.Split(s, "\n")

	paths := []string{}
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, filterPrefix) {
			paths = append(paths, line)
		}
	}
	return paths
}

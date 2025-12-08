// Package git provides git repository integration for detecting changed files.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInfo holds git metadata for a run
type GitInfo struct {
	InGitRepo    bool     `json:"inGitRepo"`
	RepoRoot     string   `json:"projectRoot"`
	Mode         string   `json:"mode"` // "staged", "staged_unstaged", "ref"
	Ref          string   `json:"ref"`  // reference used for comparison
	ChangedFiles []string `json:"changedFiles"`
}

// DetectProjectRoot detects the git repository root from current working directory
// Returns the root path and whether we're in a git repo
func DetectProjectRoot() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return ".", false
	}
	return DetectProjectRootFrom(cwd)
}

// DetectProjectRootFrom detects the git repository root starting from a specific directory
// Returns the root path and whether we're in a git repo
// If not in a git repo, returns the provided directory
func DetectProjectRootFrom(dir string) (string, bool) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir // Run git command from specified directory
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		// Not a git repo, return the directory itself
		return dir, false
	}

	root := strings.TrimSpace(buf.String())
	if root == "" {
		return dir, false
	}

	return root, true
}

// DetectChangedFiles detects changed files based on the specified mode
func DetectChangedFiles(projectRoot string, inGitRepo bool, mode string, ref string, verbose bool) GitInfo {
	info := GitInfo{
		InGitRepo:    inGitRepo,
		RepoRoot:     projectRoot,
		Mode:         mode,
		Ref:          ref,
		ChangedFiles: []string{},
	}

	if !inGitRepo {
		return info
	}

	var cmd *exec.Cmd

	switch mode {
	case "staged":
		// Only staged files
		cmd = exec.Command("git", "diff", "--cached", "--name-only")

	case "staged_unstaged":
		// Staged + unstaged files (compare against HEAD)
		cmd = exec.Command("git", "diff", "--name-only", "HEAD")

	case "ref":
		// Compare against specific ref
		cmd = exec.Command("git", "diff", "--name-only", ref)

	default:
		// Default to staged_unstaged
		cmd = exec.Command("git", "diff", "--name-only", "HEAD")
		info.Mode = "staged_unstaged"
	}

	cmd.Dir = projectRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "WARNING: git diff failed: %v\n", err)
		}
		return info
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		info.ChangedFiles = []string{}
		return info
	}

	lines := strings.Split(output, "\n")
	var files []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			files = append(files, l)
		}
	}

	info.ChangedFiles = files
	return info
}

// IsSafeDirectory checks if a directory is safe to run devpipe in.
// Returns false for system directories like /, /usr, /etc, /System, etc.
// Returns true for user directories and subdirectories of some system paths.
func IsSafeDirectory(dir string) bool {
	// Empty or relative paths are safe
	if dir == "" || dir == "." || !strings.HasPrefix(dir, "/") {
		return true
	}

	// Clean the path to normalize it (handles //, trailing slashes, etc.)
	dir = filepath.Clean(dir)

	// Special case: root directory is never safe
	if dir == "" || dir == "/" {
		return false
	}

	// Allow safe subdirectories of /usr
	if strings.HasPrefix(dir, "/usr/local/") || strings.HasPrefix(dir, "/usr/src/") {
		return true
	}

	// Allow subdirectories of /Volumes (external drives, network shares)
	if strings.HasPrefix(dir, "/Volumes/") {
		return true
	}

	// Dangerous directories where we block both the directory and all subdirectories
	strictlyDangerousDirs := []string{
		"/usr",
		"/etc",
		"/private/etc", // macOS real path for /etc
		"/bin",
		"/sbin",
		"/boot",
		"/System",
		"/Library",
		"/Applications",
		"/dev",
		"/proc",
		"/sys",
	}

	for _, dangerous := range strictlyDangerousDirs {
		if dir == dangerous || strings.HasPrefix(dir, dangerous+"/") {
			return false
		}
	}

	// Dangerous directories where we only block the top level, not subdirectories
	// (e.g., /tmp/myproject is OK, but /tmp itself is not)
	topLevelOnlyDangerousDirs := []string{
		"/var",
		"/tmp",
		"/Volumes",
	}

	for _, dangerous := range topLevelOnlyDangerousDirs {
		if dir == dangerous {
			return false
		}
	}

	return true
}

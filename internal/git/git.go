package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitInfo holds git metadata for a run
type GitInfo struct {
	InGitRepo    bool     `json:"inGitRepo"`
	RepoRoot     string   `json:"repoRoot"`
	Mode         string   `json:"mode"`         // "staged", "staged_unstaged", "ref"
	Ref          string   `json:"ref"`          // reference used for comparison
	ChangedFiles []string `json:"changedFiles"`
}

// DetectRepoRoot detects the git repository root
// Returns the root path and whether we're in a git repo
func DetectRepoRoot() (string, bool) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &bytes.Buffer{}
	
	if err := cmd.Run(); err != nil {
		// Not a git repo, use cwd
		cwd, err2 := os.Getwd()
		if err2 != nil {
			return ".", false
		}
		return cwd, false
	}
	
	root := strings.TrimSpace(buf.String())
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ".", false
		}
		return cwd, false
	}
	
	return root, true
}

// DetectChangedFiles detects changed files based on the specified mode
func DetectChangedFiles(repoRoot string, inGitRepo bool, mode string, ref string, verbose bool) GitInfo {
	info := GitInfo{
		InGitRepo:    inGitRepo,
		RepoRoot:     repoRoot,
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
	
	cmd.Dir = repoRoot
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

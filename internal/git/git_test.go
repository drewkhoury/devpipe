package git

import (
	"testing"
)

func TestDetectRepoRoot(t *testing.T) {
	// This test requires being in a git repo
	root, inGitRepo := DetectRepoRoot()
	
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Basic sanity checks
	if root == "" {
		t.Error("Expected non-empty repo root when in git repo")
	}
}

func TestDetectChangedFiles(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	tests := []struct {
		name string
		mode string
		ref  string
	}{
		{
			name: "staged mode",
			mode: "staged",
			ref:  "",
		},
		{
			name: "unstaged mode",
			mode: "unstaged",
			ref:  "",
		},
		{
			name: "staged_unstaged mode",
			mode: "staged_unstaged",
			ref:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := DetectChangedFiles(root, inGitRepo, tt.mode, tt.ref, false)
			
			// Should return a GitInfo struct
			if !info.InGitRepo {
				t.Error("Expected InGitRepo to be true")
			}
			
			if info.RepoRoot != root {
				t.Errorf("Expected RepoRoot %s, got %s", root, info.RepoRoot)
			}
		})
	}
}

func TestGitInfo(t *testing.T) {
	// Test GitInfo struct
	info := GitInfo{
		InGitRepo:    true,
		RepoRoot:     "/test/repo",
		Mode:         "staged",
		Ref:          "HEAD",
		ChangedFiles: []string{"file1.go", "file2.go"},
	}

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	if len(info.ChangedFiles) != 2 {
		t.Errorf("Expected 2 changed files, got %d", len(info.ChangedFiles))
	}

	if info.Mode != "staged" {
		t.Errorf("Expected mode 'staged', got '%s'", info.Mode)
	}
}

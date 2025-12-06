package git

import (
	"fmt"
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

func TestDetectRepoRootNotInRepo(t *testing.T) {
	// Test the case where we're not in a git repo
	// We can't easily simulate this, but we can at least verify
	// the function returns consistent results
	root, inGitRepo := DetectRepoRoot()

	// If not in a git repo, root should be empty
	if !inGitRepo && root != "" {
		t.Errorf("Expected empty root when not in git repo, got %s", root)
	}

	// If in a git repo, root should not be empty
	if inGitRepo && root == "" {
		t.Error("Expected non-empty root when in git repo")
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

func TestDetectChangedFilesNotInRepo(t *testing.T) {
	// Test when not in a git repo
	info := DetectChangedFiles("/some/path", false, "staged", "", false)

	if info.InGitRepo {
		t.Error("Expected InGitRepo to be false")
	}

	if len(info.ChangedFiles) != 0 {
		t.Errorf("Expected 0 changed files when not in repo, got %d", len(info.ChangedFiles))
	}

	if info.RepoRoot != "/some/path" {
		t.Errorf("Expected RepoRoot '/some/path', got '%s'", info.RepoRoot)
	}

	if info.Mode != "staged" {
		t.Errorf("Expected mode 'staged', got '%s'", info.Mode)
	}
}

func TestDetectChangedFilesRefMode(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test ref mode with HEAD~1 (if it exists)
	info := DetectChangedFiles(root, inGitRepo, "ref", "HEAD~1", false)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	if info.Mode != "ref" {
		t.Errorf("Expected mode 'ref', got '%s'", info.Mode)
	}

	if info.Ref != "HEAD~1" {
		t.Errorf("Expected ref 'HEAD~1', got '%s'", info.Ref)
	}

	// ChangedFiles may or may not be empty depending on repo state
	// Just verify it's initialized
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}
}

func TestDetectChangedFilesDefaultMode(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with an unknown/invalid mode - should default to staged_unstaged
	info := DetectChangedFiles(root, inGitRepo, "invalid_mode", "", false)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	if info.Mode != "staged_unstaged" {
		t.Errorf("Expected mode to default to 'staged_unstaged', got '%s'", info.Mode)
	}

	// ChangedFiles should be initialized
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}
}

func TestDetectChangedFilesVerbose(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with verbose=true - this should not panic or error
	// even if git commands fail
	info := DetectChangedFiles(root, inGitRepo, "staged", "", true)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	// Should complete successfully
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}
}

func TestDetectChangedFilesInvalidRef(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with an invalid ref - should handle gracefully
	info := DetectChangedFiles(root, inGitRepo, "ref", "nonexistent_ref_12345", false)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	// Should return empty changed files on error
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized (even if empty)")
	}
}

func TestDetectChangedFilesEmptyMode(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with empty mode - should default to staged_unstaged
	info := DetectChangedFiles(root, inGitRepo, "", "", false)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	if info.Mode != "staged_unstaged" {
		t.Errorf("Expected mode to default to 'staged_unstaged', got '%s'", info.Mode)
	}
}

func TestDetectChangedFilesWithInvalidRepoRoot(t *testing.T) {
	_, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with an invalid repo root path - git command should fail
	info := DetectChangedFiles("/nonexistent/path/to/repo", true, "staged", "", false)

	// Should handle gracefully and return empty changed files
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}

	if len(info.ChangedFiles) != 0 {
		t.Errorf("Expected 0 changed files with invalid repo root, got %d", len(info.ChangedFiles))
	}
}

func TestDetectChangedFilesWithInvalidRepoRootVerbose(t *testing.T) {
	_, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with an invalid repo root path and verbose=true
	// This should trigger the verbose error message path
	info := DetectChangedFiles("/nonexistent/path/to/repo", true, "staged", "", true)

	// Should handle gracefully and return empty changed files
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}

	if len(info.ChangedFiles) != 0 {
		t.Errorf("Expected 0 changed files with invalid repo root, got %d", len(info.ChangedFiles))
	}
}

func TestGitInfoStructFields(t *testing.T) {
	// Test all GitInfo struct fields are properly set
	info := GitInfo{
		InGitRepo:    false,
		RepoRoot:     "",
		Mode:         "",
		Ref:          "",
		ChangedFiles: nil,
	}

	if info.InGitRepo {
		t.Error("Expected InGitRepo to be false")
	}

	if info.RepoRoot != "" {
		t.Errorf("Expected empty RepoRoot, got '%s'", info.RepoRoot)
	}

	if info.Mode != "" {
		t.Errorf("Expected empty Mode, got '%s'", info.Mode)
	}

	if info.Ref != "" {
		t.Errorf("Expected empty Ref, got '%s'", info.Ref)
	}

	if info.ChangedFiles != nil {
		t.Errorf("Expected nil ChangedFiles, got %v", info.ChangedFiles)
	}
}

func TestDetectChangedFilesAllModes(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test all valid modes comprehensively
	modes := []struct {
		mode         string
		expectedMode string
		ref          string
	}{
		{"staged", "staged", ""},
		{"staged_unstaged", "staged_unstaged", ""},
		{"ref", "ref", "HEAD"},
		{"unknown", "staged_unstaged", ""}, // Should default
		{"", "staged_unstaged", ""},        // Empty should default
	}

	for _, tc := range modes {
		t.Run("mode_"+tc.mode, func(t *testing.T) {
			info := DetectChangedFiles(root, inGitRepo, tc.mode, tc.ref, false)

			if info.Mode != tc.expectedMode {
				t.Errorf("Mode %s: expected mode '%s', got '%s'", tc.mode, tc.expectedMode, info.Mode)
			}

			if info.InGitRepo != inGitRepo {
				t.Errorf("Mode %s: expected InGitRepo %v, got %v", tc.mode, inGitRepo, info.InGitRepo)
			}

			if info.RepoRoot != root {
				t.Errorf("Mode %s: expected RepoRoot '%s', got '%s'", tc.mode, root, info.RepoRoot)
			}

			if info.ChangedFiles == nil {
				t.Errorf("Mode %s: expected ChangedFiles to be initialized", tc.mode)
			}
		})
	}
}

func TestDetectChangedFilesRefWithEmptyRef(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test ref mode with empty ref string - should fail gracefully
	info := DetectChangedFiles(root, inGitRepo, "ref", "", false)

	if !info.InGitRepo {
		t.Error("Expected InGitRepo to be true")
	}

	if info.Mode != "ref" {
		t.Errorf("Expected mode 'ref', got '%s'", info.Mode)
	}

	// Empty ref should cause git diff to fail, returning empty files
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be initialized")
	}
}

func TestDetectChangedFilesMultipleRefs(t *testing.T) {
	root, inGitRepo := DetectRepoRoot()
	if !inGitRepo {
		t.Skip("Skipping git test: not in a git repository")
		return
	}

	// Test with various ref formats
	refs := []string{
		"HEAD",
		"HEAD~1",
		"HEAD~2",
		"origin/main",
		"main",
	}

	for _, ref := range refs {
		t.Run("ref_"+ref, func(t *testing.T) {
			info := DetectChangedFiles(root, inGitRepo, "ref", ref, false)

			if info.Mode != "ref" {
				t.Errorf("Ref %s: expected mode 'ref', got '%s'", ref, info.Mode)
			}

			if info.Ref != ref {
				t.Errorf("Ref %s: expected ref '%s', got '%s'", ref, ref, info.Ref)
			}

			// ChangedFiles should be initialized (may be empty if ref doesn't exist)
			if info.ChangedFiles == nil {
				t.Errorf("Ref %s: expected ChangedFiles to be initialized", ref)
			}
		})
	}
}

func TestDetectRepoRootConsistency(t *testing.T) {
	// Test that DetectRepoRoot returns consistent results when called multiple times
	root1, inRepo1 := DetectRepoRoot()
	root2, inRepo2 := DetectRepoRoot()

	if root1 != root2 {
		t.Errorf("DetectRepoRoot returned inconsistent roots: '%s' vs '%s'", root1, root2)
	}

	if inRepo1 != inRepo2 {
		t.Errorf("DetectRepoRoot returned inconsistent inGitRepo values: %v vs %v", inRepo1, inRepo2)
	}
}

func TestGitInfoWithEmptyChangedFiles(t *testing.T) {
	// Test GitInfo with explicitly empty ChangedFiles slice
	info := GitInfo{
		InGitRepo:    true,
		RepoRoot:     "/test/repo",
		Mode:         "staged",
		Ref:          "",
		ChangedFiles: []string{},
	}

	if len(info.ChangedFiles) != 0 {
		t.Errorf("Expected 0 changed files, got %d", len(info.ChangedFiles))
	}

	// Verify it's not nil
	if info.ChangedFiles == nil {
		t.Error("Expected ChangedFiles to be non-nil empty slice")
	}
}

func TestGitInfoWithManyChangedFiles(t *testing.T) {
	// Test GitInfo with many changed files
	files := make([]string, 100)
	for i := 0; i < 100; i++ {
		files[i] = fmt.Sprintf("file%d.go", i)
	}

	info := GitInfo{
		InGitRepo:    true,
		RepoRoot:     "/test/repo",
		Mode:         "staged_unstaged",
		Ref:          "HEAD",
		ChangedFiles: files,
	}

	if len(info.ChangedFiles) != 100 {
		t.Errorf("Expected 100 changed files, got %d", len(info.ChangedFiles))
	}

	// Verify first and last files
	if info.ChangedFiles[0] != "file0.go" {
		t.Errorf("Expected first file 'file0.go', got '%s'", info.ChangedFiles[0])
	}

	if info.ChangedFiles[99] != "file99.go" {
		t.Errorf("Expected last file 'file99.go', got '%s'", info.ChangedFiles[99])
	}
}

func TestDetectChangedFilesNotInRepoWithDifferentModes(t *testing.T) {
	// Test all modes when not in a git repo - should all return empty
	modes := []string{"staged", "staged_unstaged", "ref", "invalid", ""}

	for _, mode := range modes {
		t.Run("notInRepo_"+mode, func(t *testing.T) {
			info := DetectChangedFiles("/some/path", false, mode, "HEAD", false)

			if info.InGitRepo {
				t.Error("Expected InGitRepo to be false")
			}

			if len(info.ChangedFiles) != 0 {
				t.Errorf("Mode %s: expected 0 changed files when not in repo, got %d", mode, len(info.ChangedFiles))
			}

			// Mode should be preserved as-is when not in repo
			if info.Mode != mode {
				t.Errorf("Mode %s: expected mode '%s', got '%s'", mode, mode, info.Mode)
			}
		})
	}
}

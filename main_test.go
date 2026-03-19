package main

import (
	"errors"
	"testing"
)

type mockGitCommander struct {
	worktreeListOutput string
	worktreeListErr    error
	revParseOutput     string
	revParseErr        error
	worktreeRemoveErr  error
	worktreeAddErr     error
	worktreeAddNewErr  error
}

func (m *mockGitCommander) worktreeList() (string, error) {
	return m.worktreeListOutput, m.worktreeListErr
}

func (m *mockGitCommander) revParse(showToplevel bool) (string, error) {
	return m.revParseOutput, m.revParseErr
}

func (m *mockGitCommander) worktreeRemove(path string, force bool) error {
	return m.worktreeRemoveErr
}

func (m *mockGitCommander) worktreeAdd(path, branch string) error {
	return m.worktreeAddErr
}

func (m *mockGitCommander) worktreeAddNew(path, branch string) error {
	return m.worktreeAddNewErr
}

func TestExtractBranch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "branch in brackets",
			input:    "/path/to/worktree abc123 [feature-branch]",
			expected: "feature-branch",
		},
		{
			name:     "detached HEAD in parentheses",
			input:    "/path/to/worktree abc123 (detached at abc1234)",
			expected: "detached at abc1234",
		},
		{
			name:     "no brackets or parentheses",
			input:    "/path/to/worktree abc123",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only opening bracket",
			input:    "[unclosed",
			expected: "",
		},
		{
			name:     "only opening paren",
			input:    "(unclosed",
			expected: "",
		},
		{
			name:     "multiple brackets returns first",
			input:    "[first] something [second]",
			expected: "first",
		},
		{
			name:     "branch with slashes",
			input:    "/path/to/worktree abc123 [feature/my-branch]",
			expected: "feature/my-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBranch(tt.input)
			if result != tt.expected {
				t.Errorf("extractBranch(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetWorktreesSorted(t *testing.T) {
	tests := []struct {
		name           string
		worktreeOutput string
		gitRoot        string
		wantErr        bool
		expectedCount  int
		expectedFirst  string
	}{
		{
			name: "single worktree with branch",
			worktreeOutput: `/home/user/project-main abc123 [main]
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user",
			wantErr:       false,
			expectedCount: 2,
			expectedFirst: "/home/user/project-main",
		},
		{
			name:           "worktree list error",
			worktreeOutput: "",
			gitRoot:        "/home/user",
			wantErr:        true,
		},
		{
			name:           "empty worktree list",
			worktreeOutput: "",
			gitRoot:        "/home/user",
			wantErr:        false,
			expectedCount:  0,
		},
		{
			name: "worktree with detached HEAD",
			worktreeOutput: `/home/user/project-main abc123 (detached at abc1234)
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user",
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name: "worktree without branch",
			worktreeOutput: `/home/user/project-main abc123
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user",
			wantErr:       false,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := &mockGitCommander{
				worktreeListOutput: tt.worktreeOutput,
				revParseOutput:     tt.gitRoot,
			}

			result, err := getWorktreesSorted(commander)

			if (err != nil) != tt.wantErr {
				t.Errorf("getWorktreesSorted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("getWorktreesSorted() returned %d items, want %d", len(result), tt.expectedCount)
			}

			if tt.expectedCount > 0 && result[0].AbsPath != tt.expectedFirst {
				t.Errorf("First worktree AbsPath = %q, want %q", result[0].AbsPath, tt.expectedFirst)
			}
		})
	}
}

func TestFindWorktreePathForBranch(t *testing.T) {
	tests := []struct {
		name      string
		branch    string
		commander *mockGitCommander
		wantPath  string
		wantErr   bool
	}{
		{
			name:   "branch found",
			branch: "feature-test",
			commander: &mockGitCommander{
				worktreeListOutput: `/home/user/project-main abc123 [main]
/home/user/project-feature def456 [feature-test]`,
				revParseOutput: "/home/user",
			},
			wantPath: "/home/user/project-feature",
			wantErr:  false,
		},
		{
			name:   "branch not found",
			branch: "nonexistent",
			commander: &mockGitCommander{
				worktreeListOutput: `/home/user/project-main abc123 [main]`,
				revParseOutput:     "/home/user",
			},
			wantPath: "",
			wantErr:  true,
		},
		{
			name:      "empty branch",
			branch:    "",
			commander: &mockGitCommander{},
			wantPath:  "",
			wantErr:   true,
		},
		{
			name:   "worktree list error",
			branch: "feature-test",
			commander: &mockGitCommander{
				worktreeListErr: errors.New("git not found"),
				revParseOutput:  "/home/user",
			},
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := findWorktreePathForBranch(tt.branch, tt.commander)

			if (err != nil) != tt.wantErr {
				t.Errorf("findWorktreePathForBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if path != tt.wantPath {
				t.Errorf("findWorktreePathForBranch() = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

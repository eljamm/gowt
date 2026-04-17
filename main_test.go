package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"gwt/internal/app"
	"gwt/internal/git"
	"gwt/internal/tui"
)

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
			result := app.ExtractBranch(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractBranch(%q) = %q, want %q", tt.input, result, tt.expected)
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
			name: "cwd worktree sorted first",
			worktreeOutput: `/home/user/project-main abc123 [main]
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user/project-main",
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
			wantErr:        true,
		},
		{
			name: "worktree with detached HEAD",
			worktreeOutput: `/home/user/project-main abc123 (detached at abc1234)
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user/project-main",
			wantErr:       false,
			expectedCount: 2,
			expectedFirst: "/home/user/project-main",
		},
		{
			name: "worktree without branch",
			worktreeOutput: `/home/user/project-main abc123
/home/user/project-feature def456 [feature-test]`,
			gitRoot:       "/home/user/project-main",
			wantErr:       false,
			expectedCount: 2,
			expectedFirst: "/home/user/project-main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := &git.MockCommander{
				WorktreeListOutput: tt.worktreeOutput,
				RevParseOutput:     tt.gitRoot,
			}

			result, err := app.GetWorktreesSorted(commander)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorktreesSorted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf(
					"GetWorktreesSorted() returned %d items, want %d",
					len(result),
					tt.expectedCount,
				)
			}

			if tt.expectedCount > 0 && result[0].AbsPath != tt.expectedFirst {
				t.Errorf(
					"First worktree AbsPath = %q, want %q",
					result[0].AbsPath,
					tt.expectedFirst,
				)
			}
		})
	}
}

func TestFuzzyMatchPositions(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		target  string
		wantNil bool
		wantLen int
	}{
		{
			name:    "exact match",
			query:   "main",
			target:  "main [abc123] [main]",
			wantNil: false,
			wantLen: 4,
		},
		{
			name:    "partial match",
			query:   "mn",
			target:  "main [abc123] [main]",
			wantNil: false,
			wantLen: 2,
		},
		{
			name:    "no match",
			query:   "xyz",
			target:  "main [abc123] [main]",
			wantNil: true,
		},
		{
			name:    "empty query",
			query:   "",
			target:  "main [abc123] [main]",
			wantNil: true,
		},
		{
			name:    "case insensitive",
			query:   "MAIN",
			target:  "main [abc123] [main]",
			wantNil: false,
			wantLen: 4,
		},
		{
			name:    "non-sequential fails",
			query:   "mnx",
			target:  "main [abc123] [main]",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.FuzzyMatchPositions(tt.query, tt.target)
			if tt.wantNil {
				if result != nil {
					t.Errorf(
						"FuzzyMatchPositions(%q, %q) = %v, want nil",
						tt.query,
						tt.target,
						result,
					)
				}
				return
			}
			if result == nil {
				t.Errorf(
					"FuzzyMatchPositions(%q, %q) = nil, want %d positions",
					tt.query,
					tt.target,
					tt.wantLen,
				)
			}
			if len(result) != tt.wantLen {
				t.Errorf(
					"FuzzyMatchPositions(%q, %q) returned %d positions, want %d",
					tt.query,
					tt.target,
					len(result),
					tt.wantLen,
				)
			}
		})
	}
}

func TestFindWorktreePathForBranch(t *testing.T) {
	tests := []struct {
		name           string
		branch         string
		worktreeOutput string
		gitRoot        string
		wantErr        bool
		wantPath       string
	}{
		{
			name:   "branch found",
			branch: "feature-test",
			worktreeOutput: `/home/user/project-main abc123 [main]
/home/user/project-feature def456 [feature-test]`,
			gitRoot:  "/home/user/project-main",
			wantErr:  false,
			wantPath: "/home/user/project-feature",
		},
		{
			name:   "branch not found",
			branch: "nonexistent",
			worktreeOutput: `/home/user/project-main abc123 [main]
/home/user/project-feature def456 [feature-test]`,
			gitRoot: "/home/user/project-main",
			wantErr: true,
		},
		{
			name:           "empty worktrees",
			branch:         "main",
			worktreeOutput: "",
			gitRoot:        "/home/user/project-main",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := &git.MockCommander{
				WorktreeListOutput: tt.worktreeOutput,
				RevParseOutput:     tt.gitRoot,
			}

			path, err := app.FindWorktreePathForBranch(tt.branch, commander)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"FindWorktreePathForBranch(%q) error = %v, wantErr %v",
					tt.branch,
					err,
					tt.wantErr,
				)
				return
			}

			if !tt.wantErr && path != tt.wantPath {
				t.Errorf(
					"FindWorktreePathForBranch(%q) = %q, want %q",
					tt.branch,
					path,
					tt.wantPath,
				)
			}
		})
	}
}

func TestStateTransitions(t *testing.T) {
	t.Run("NormalState navigation", func(t *testing.T) {
		state := tui.NormalState{}

		_, _, req := state.HandleKey(tui.KeyEvent{Rune: 'g'})
		if !req.NavTop {
			t.Error("expected NavTop for 'g'")
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'G'})
		if !req.NavBottom {
			t.Error("expected NavBottom for 'G'")
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'j'})
		if req.NavDelta != 1 {
			t.Errorf("NavDelta = %d, want 1 for 'j'", req.NavDelta)
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'k'})
		if req.NavDelta != -1 {
			t.Errorf("NavDelta = %d, want -1 for 'k'", req.NavDelta)
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'J'})
		if req.NavDelta != 1 {
			t.Errorf("NavDelta = %d, want 1 for 'J'", req.NavDelta)
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'K'})
		if req.NavDelta != -1 {
			t.Errorf("NavDelta = %d, want -1 for 'K'", req.NavDelta)
		}
	})

	t.Run("NormalState with count prefix", func(t *testing.T) {
		state := tui.NormalState{Count: 3}

		_, _, req := state.HandleKey(tui.KeyEvent{Rune: 'k'})
		if req.NavDelta != -3 {
			t.Errorf("NavDelta = %d, want -3 for '3k'", req.NavDelta)
		}

		state = tui.NormalState{Count: 3}
		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'j'})
		if req.NavDelta != 3 {
			t.Errorf("NavDelta = %d, want 3 for '3j'", req.NavDelta)
		}

		state = tui.NormalState{Count: 20}
		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'g'})
		if req.NavTarget != 19 {
			t.Errorf("NavTarget = %d, want 19 for '20g'", req.NavTarget)
		}

		state = tui.NormalState{Count: 1}
		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'G'})
		if req.NavTarget != 0 {
			t.Errorf("NavTarget = %d, want 0 for '1G'", req.NavTarget)
		}

		state = tui.NormalState{Count: 0}
		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'g'})
		if !req.NavTop {
			t.Error("expected NavTop for plain 'g'")
		}

		state = tui.NormalState{Count: 0}
		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'G'})
		if !req.NavBottom {
			t.Error("expected NavBottom for plain 'G'")
		}
	})

	t.Run("NormalState digit accumulation", func(t *testing.T) {
		state := tui.NormalState{}

		nextState, _, _ := state.HandleKey(tui.KeyEvent{Rune: '3'})
		state = nextState.(tui.NormalState)
		if state.Count != 3 {
			t.Errorf("Count = %d, want 3", state.Count)
		}

		nextState, _, _ = state.HandleKey(tui.KeyEvent{Rune: '2'})
		state = nextState.(tui.NormalState)
		if state.Count != 32 {
			t.Errorf("Count = %d, want 32", state.Count)
		}

		nextState, _, _ = state.HandleKey(tui.KeyEvent{Rune: '0'})
		state = nextState.(tui.NormalState)
		if state.Count != 320 {
			t.Errorf("Count = %d, want 320", state.Count)
		}

		_, _, _ = state.HandleKey(tui.KeyEvent{Rune: 'i'})
	})

	t.Run("NormalState mode changes", func(t *testing.T) {
		state := tui.NormalState{}

		nextState, _, req := state.HandleKey(tui.KeyEvent{Rune: 'i'})
		if _, isInsert := nextState.(tui.InsertState); !isInsert {
			t.Error("expected InsertState for 'i'")
		}
		if req.Query != "" {
			t.Error("expected empty query on mode change")
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Rune: 'q'})
		if req.ModeIndicator != " ? " {
			t.Errorf("ModeIndicator = %q, want ' ? '", req.ModeIndicator)
		}

		_, _, req = state.HandleKey(tui.KeyEvent{Key: tcell.KeyEsc})
		if req.ModeIndicator != " ? " {
			t.Errorf("ModeIndicator = %q, want ' ? ' for ESC", req.ModeIndicator)
		}

		_, action, _ := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEnter})
		if action != tui.ActionSelect {
			t.Error("expected ActionSelect for Enter")
		}

		_, action, _ = state.HandleKey(tui.KeyEvent{Key: tcell.KeyCtrlC})
		if action != tui.ActionQuit {
			t.Error("expected ActionQuit for Ctrl+C")
		}
	})

	t.Run("InsertState typing", func(t *testing.T) {
		state := tui.InsertState{Query: "feat"}

		_, _, req := state.HandleKey(tui.KeyEvent{Rune: 'h'})
		if req.Query != "feath" {
			t.Errorf("query = %q, want %q", req.Query, "feath")
		}
	})

	t.Run("InsertState backspace", func(t *testing.T) {
		state := tui.InsertState{Query: "test"}

		_, _, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyBackspace})
		if req.Query != "tes" {
			t.Errorf("query = %q, want %q", req.Query, "tes")
		}
	})

	t.Run("InsertState backspace2", func(t *testing.T) {
		state := tui.InsertState{Query: "tes"}

		_, _, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyBackspace2})
		if req.Query != "te" {
			t.Errorf("query = %q, want %q", req.Query, "te")
		}
	})

	t.Run("InsertState escape with query clears query", func(t *testing.T) {
		state := tui.InsertState{Query: "some query"}

		nextState, _, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEsc})
		if req.Query != "" {
			t.Errorf("query = %q, want empty after ESC", req.Query)
		}

		_, isInsert := nextState.(tui.InsertState)
		if !isInsert {
			t.Error("expected InsertState after ESC (stays in insert to clear query)")
		}
		if nextState.(tui.InsertState).Query != "" {
			t.Error("expected query cleared in state")
		}
	})

	t.Run("InsertState escape with empty query", func(t *testing.T) {
		state := tui.InsertState{Query: ""}

		_, _, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEsc})
		if req.ModeIndicator != " N " {
			t.Errorf("ModeIndicator = %q, want ' N '", req.ModeIndicator)
		}
	})

	t.Run("InsertState enter selects", func(t *testing.T) {
		state := tui.InsertState{Query: "test"}

		_, action, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEnter})
		if action != tui.ActionSelect {
			t.Error("expected ActionSelect for Enter")
		}
		if req.Query != "test" {
			t.Errorf("query = %q in request", req.Query)
		}
	})

	t.Run("ConfirmQuitState yes", func(t *testing.T) {
		state := tui.ConfirmQuitState{}

		_, action, _ := state.HandleKey(tui.KeyEvent{Rune: 'y'})
		if action != tui.ActionQuit {
			t.Error("expected ActionQuit for 'y'")
		}
	})

	t.Run("ConfirmQuitState no returns Normal", func(t *testing.T) {
		state := tui.ConfirmQuitState{}

		_, action, req := state.HandleKey(tui.KeyEvent{Rune: 'n'})
		if action != tui.ActionDraw {
			t.Error("expected ActionDraw for 'n'")
		}
		if req.ModeIndicator != " N " {
			t.Errorf("ModeIndicator = %q, want ' N '", req.ModeIndicator)
		}
	})

	t.Run("ConfirmQuitState enter quits", func(t *testing.T) {
		state := tui.ConfirmQuitState{}

		_, action, _ := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEnter})
		if action != tui.ActionQuit {
			t.Error("expected ActionQuit for Enter in confirm")
		}
	})

	t.Run("ConfirmQuitState escape returns Normal", func(t *testing.T) {
		state := tui.ConfirmQuitState{}

		_, _, req := state.HandleKey(tui.KeyEvent{Key: tcell.KeyEsc})
		if req.ModeIndicator != " N " {
			t.Errorf("ModeIndicator = %q, want ' N '", req.ModeIndicator)
		}
	})
}

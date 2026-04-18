package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var adjectives = []string{
	"happy", "brave", "clever", "gentle", "swift",
	"bright", "calm", "eager", "fierce", "kind",
	"lively", "mighty", "noble", "proud", "quick",
	"wise", "zany", "ample", "bold", "cool",
	"dear", "earnest", "fair", "grace", "hard",
	"keen", "light", "merry", "new", "open",
	"peace", "pure", "rich", "sure", "top",
	"vast", "wild", "young", "alert", "alive",
	"busy", "cute", "fine", "free", "glad",
	"warm", "rare", "real", "solid", "tame",
}

var animals = []string{
	"fox", "owl", "bear", "wolf", "hawk",
	"deer", "lion", "tiger", "eagle", "crane",
	"heron", "otter", "panda", "rabbit", "swan",
	"zebra", "badger", "coyote", "dolphin", "elk",
	"flamingo", "gator", "hyena", "iguana", "jaguar",
	"koala", "lemur", "marmoset", "newt", "ocelot",
	"peacock", "quail", "rhino", "sloth", "tapir",
	"uxv", "viper", "wallaby", "xerus", "yak",
	"zebu", "alpaca", "baboon", "capybara", "dingo",
	"emu", "ferret", "gibbon", "hedgehog", "impala",
}

type worktree struct {
	name string
	path string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: demo-worktrees.sh <setup|clean>")
		fmt.Println("  setup - create 50 demo worktrees")
		fmt.Println("  clean - remove all demo worktrees")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		setup()
	case "clean":
		clean()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func setup() {
	names := generateNames(50)

	root, err := findRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding git root: %v\n", err)
		os.Exit(1)
	}

	wtDir := filepath.Join(root, ".wt")
	if err := os.MkdirAll(wtDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .wt directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Creating %d demo worktrees in %s...\n", len(names), wtDir)

	for i, name := range names {
		path := filepath.Join(wtDir, name)

		if _, err := os.Stat(path); err == nil {
			fmt.Printf("[%d/%d] Skipping (path exists): %s\n", i+1, len(names), name)
			continue
		}

		if branchExists(root, name) {
			if err := deleteBranch(root, name); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting branch %s: %v\n", name, err)
			}
		}

		if err := createWorktree(root, name, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktree %s: %v\n", name, err)
			continue
		}
		fmt.Printf("[%d/%d] Created: %s\n", i+1, len(names), name)
	}

	fmt.Println("\nDone! Run './gwt' to see the demo worktrees.")
	fmt.Println("Use './scripts/demo-worktrees.go clean' to remove them.")
}

func clean() {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding git root: %v\n", err)
		os.Exit(1)
	}

	worktrees, err := listWorktrees(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
		os.Exit(1)
	}

	if len(worktrees) <= 1 {
		fmt.Println("No demo worktrees to remove.")
		return
	}

	fmt.Printf("Removing %d worktrees...\n", len(worktrees)-1)

	for _, wt := range worktrees {
		if wt.path == root {
			continue
		}
		if err := removeWorktree(root, wt.path); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", wt.name, err)
			continue
		}
		fmt.Printf("Removed: %s\n", wt.name)
	}

	fmt.Println("\nDone! All demo worktrees removed.")
}

func generateNames(count int) []string {
	names := make([]string, 0, count)

	shuffledAdjs := make([]string, len(adjectives))
	copy(shuffledAdjs, adjectives)
	rand.Shuffle(len(shuffledAdjs), func(i, j int) {
		shuffledAdjs[i], shuffledAdjs[j] = shuffledAdjs[j], shuffledAdjs[i]
	})

	shuffledAnimals := make([]string, len(animals))
	copy(shuffledAnimals, animals)
	rand.Shuffle(len(shuffledAnimals), func(i, j int) {
		shuffledAnimals[i], shuffledAnimals[j] = shuffledAnimals[j], shuffledAnimals[i]
	})

	for i := 0; i < count && i < len(shuffledAnimals); i++ {
		name := shuffledAdjs[i] + "-" + shuffledAnimals[i]
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func findRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func listWorktrees(root string) ([]worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	worktrees := make([]worktree, 0)

	var (
		path   string
		branch string
	)

	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "worktree "); ok {
			path = after
		} else if after0, ok0 := strings.CutPrefix(line, "branch "); ok0 {
			branch = after0
			branch = strings.TrimPrefix(branch, "refs/heads/")
			worktrees = append(worktrees, worktree{name: branch, path: path})
		}
	}

	return worktrees, nil
}

func createWorktree(root, branch, path string) error {
	cmd := exec.Command("git", "worktree", "add", path, "-b", branch)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "already exists") {
			return fmt.Errorf("%s: %s", out, err)
		}
	}
	return nil
}

func removeWorktree(root, path string) error {
	cmd := exec.Command("git", "worktree", "remove", path, "--force")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", out, err)
	}
	return nil
}

func branchExists(root, name string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	cmd.Dir = root
	return cmd.Run() == nil
}

func deleteBranch(root, name string) error {
	cmd := exec.Command("git", "branch", "-D", name)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", out, err)
	}
	return nil
}

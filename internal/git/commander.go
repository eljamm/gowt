package git

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Commander interface {
	WorktreeList() (string, error)
	RevParse(showToplevel bool) (string, error)
	GitCommonDir() (string, error)
	WorktreeRemove(path string, force bool) error
	WorktreeAdd(path, branch string) error
	WorktreeAddNew(path, branch string) error
	BranchDelete(branch string, force bool) error
}

type RealCommander struct{}

func (r *RealCommander) WorktreeList() (string, error) {
	out, err := exec.Command("git", "worktree", "list").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (r *RealCommander) RevParse(showToplevel bool) (string, error) {
	args := []string{"rev-parse"}
	if showToplevel {
		args = append(args, "--show-toplevel")
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealCommander) GitCommonDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealCommander) WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force", path)
	} else {
		args = append(args, path)
	}
	c := exec.Command("git", args...)
	c.Stderr = os.Stderr
	return c.Run()
}

func (r *RealCommander) WorktreeAdd(path, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", path, branch)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *RealCommander) WorktreeAddNew(path, branch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branch, path)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *RealCommander) BranchDelete(branch string, force bool) error {
	args := []string{"branch"}
	if force {
		args = append(args, "-D", branch)
	} else {
		args = append(args, "-d", branch)
	}
	c := exec.Command("git", args...)
	c.Stderr = os.Stderr
	return c.Run()
}

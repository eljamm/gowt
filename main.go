package main

import (
	"os"

	"github.com/spf13/cobra"

	"gwt/internal/app"
	"gwt/internal/git"
	"gwt/internal/tui"
)

var defaultCommander git.Commander = &git.RealCommander{}

func main() {
	app.TUIColors = tui.DefaultTUIColors()
	app.SelectWorktreeTUI = tui.SelectWorktreeTUI

	rootCmd := &cobra.Command{
		Use:   "gwt",
		Short: "Git Worktree Manager",
		Run:   func(cmd *cobra.Command, args []string) { app.RunJump(cmd, args, defaultCommander) },
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "add [branch]",
		Short: "Create a worktree from a branch",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) { app.RunAdd(cmd, args, defaultCommander) },
	})

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Interactively remove a worktree",
		Run:   func(cmd *cobra.Command, args []string) { app.RunRemove(cmd, args, defaultCommander) },
	}
	removeCmd.Flags().BoolP("force", "f", false, "Force removal")
	rootCmd.AddCommand(removeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

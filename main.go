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

	appCmd := &cobra.Command{
		Use:   "gwt",
		Short: "Git Worktree Manager",
		Run:   func(cmd *cobra.Command, args []string) { app.RunJump(cmd, args, defaultCommander) },
	}

	appCmd.AddCommand(&cobra.Command{
		Use:   "add [wt-name] [branch]",
		Short: "Create a worktree",
		Args:  cobra.RangeArgs(0, 2),
		Run:   func(cmd *cobra.Command, args []string) { app.RunAdd(cmd, args, defaultCommander) },
	})

	removeCmd := &cobra.Command{
		Use:   "remove [worktree-name...]",
		Short: "Remove one or more worktrees",
		Args:  cobra.ArbitraryArgs,
		Run:   func(cmd *cobra.Command, args []string) { app.RunRemove(cmd, args, defaultCommander) },
	}
	removeCmd.Flags().BoolP("force", "f", false, "Force removal")
	appCmd.AddCommand(removeCmd)

	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Jump to git root (toplevel worktree)",
		Run:   func(cmd *cobra.Command, args []string) { app.RunRoot(cmd, args, defaultCommander) },
	}
	appCmd.AddCommand(rootCmd)

	homeCmd := &cobra.Command{
		Use:   "home",
		Short: "Jump to main worktree",
		Run:   func(cmd *cobra.Command, args []string) { app.RunHome(cmd, args, defaultCommander) },
	}
	appCmd.AddCommand(homeCmd)

	purgeCmd := &cobra.Command{
		Use:   "purge [worktree-name...]",
		Short: "Remove worktrees and delete their branches",
		Args:  cobra.ArbitraryArgs,
		Run:   func(cmd *cobra.Command, args []string) { app.RunPurge(cmd, args, defaultCommander) },
	}
	purgeCmd.Flags().BoolP("force", "f", false, "Force removal and branch deletion")
	appCmd.AddCommand(purgeCmd)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}
	mainCmd := &cobra.Command{
		Use:   "main [path]",
		Short: "Get/set main worktree path",
		Args:  cobra.RangeArgs(0, 1),
		Run:   func(cmd *cobra.Command, args []string) { app.RunConfig(cmd, args, defaultCommander) },
	}
	configCmd.AddCommand(mainCmd)
	appCmd.AddCommand(configCmd)

	if err := appCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

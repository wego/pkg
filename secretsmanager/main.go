package main

import (
	"github.com/spf13/cobra"
	"github.com/wego/pkg/secretsmanager/cmd"
)

func main() {
	var rootCmd = cobra.Command{}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(cmd.UpdateCmd(cmd.UpdateCmdConfig{}))
	rootCmd.AddCommand(cmd.BackupCmd())

	rootCmd.Execute()
}

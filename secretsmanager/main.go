package main

import (
	"github.com/spf13/cobra"
	"github.com/wego/pkg/secretsmanager/cmd"
)

const defaultAWSRegion = "ap-southeast-1"

func main() {
	var rootCmd = cobra.Command{}

	rootCmd.PersistentFlags().StringP("secret-id", "s", "", "Secret id to be updated")
	rootCmd.PersistentFlags().StringP("aws-profile", "p", "", "Specify the aws sso profile")

	rootCmd.AddCommand(cmd.UpdateCmd())
	rootCmd.Execute()
}

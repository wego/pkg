package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// BackupCmd backup secret on AWS Secrets Manager
func BackupCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "backup",
		Short: "Backup secret on AWS Secrets Manager",
		Run: func(cmd *cobra.Command, args []string) {
			flag.Parse()

			secretID, _ := cmd.Flags().GetString("secret-id")
			secretID = strings.TrimSpace(secretID)
			if len(secretID) == 0 {
				cmd.Help()
				os.Exit(1)
			}

			awsProfile, _ := cmd.Flags().GetString("aws-profile")
			awsProfile = strings.TrimSpace(awsProfile)
			if len(awsProfile) == 0 {
				cmd.Help()
				os.Exit(1)
			}

			secret, err := retrieveSecret(secretID, awsProfile)
			if err != nil {
				log.Fatal(err)
			}

			backupFileName := fmt.Sprintf("%s-%s", *secret.Name, secret.CreatedDate.UTC().Format(time.RFC3339))
			if err := ioutil.WriteFile(backupFileName, []byte(*secret.SecretString), 0644); err != nil {
				log.Fatal(err)
			}

			log.Println("Backup to file", backupFileName)
		},
	}

	return &cmd
}
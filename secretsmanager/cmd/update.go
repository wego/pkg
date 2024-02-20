package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/browser"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

const (
	defaultEditor = "vim"
)

// UpdateCmdConfig ...
type UpdateCmdConfig struct {
	Validate func(secretString string) error
}

// UpdateCmd update secret on AWS Secrets Manager
func UpdateCmd(c UpdateCmdConfig) *cobra.Command {
	cmd := cobra.Command{
		Use:   "update",
		Short: "Update secret on AWS Secrets Manager",
		Run: func(cmd *cobra.Command, _ []string) {
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

			newSecretString := openSecretInEditor(secret)
			if *secret.SecretString == newSecretString {
				log.Println("No changes in the secret, aborting...")
				os.Exit(0)
			}

			if !confirmDiffs(*secret.SecretString, newSecretString) {
				os.Exit(0)
			}

			if c.Validate != nil {
				if err := c.Validate(newSecretString); err != nil {
					log.Fatal(err)
				}
			}

			if err := checkLatestSecretVersion(secret, awsProfile); err != nil {
				log.Fatal(err)
			}

			log.Println("Updating secret to AWS...")
			if _, err := updateSecret(*secret.ARN, awsProfile, newSecretString); err != nil {
				log.Fatal(err)
			}
			log.Println("Updated successfully.")
		},
	}

	cmd.PersistentFlags().StringP("secret-id", "s", "", "Secret id to be updated")
	cmd.PersistentFlags().StringP("aws-profile", "p", "", "Specify the aws sso profile")
	cmd.CompletionOptions.DisableDefaultCmd = true

	return &cmd
}

func openSecretInEditor(secret *secretsmanager.GetSecretValueOutput) (updated string) {
	tempFileName := fmt.Sprintf("%s-%s", *secret.Name, secret.CreatedDate.UTC().Format(time.RFC3339))
	f, err := ioutil.TempFile("", tempFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	editor := os.Getenv("EDITOR")
	if len(editor) == 0 {
		editor = defaultEditor
	}

	path, err := exec.LookPath(editor)
	if err != nil {
		log.Fatalf("Error %s while looking up for %s!!", path, editor)
	}

	editCmdArgs := []string{}
	switch editor {
	case "code", "goland", "idea":
		editCmdArgs = append(editCmdArgs, "--wait")
	}

	if _, err := f.WriteString(*secret.SecretString); err != nil {
		log.Fatal(err)
	}
	editCmdArgs = append(editCmdArgs, f.Name())

	editCmd := exec.Command(path, editCmdArgs...)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	if err := editCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := editCmd.Wait(); err != nil {
		log.Fatal(err)
	}

	newSecret, err := ioutil.ReadFile(f.Name())
	if err != nil {
		log.Fatal(err)
	}

	return string(newSecret)
}

func renderHTML(diffs []diffmatchpatch.Diff) string {
	var output bytes.Buffer

	tmpl := template.Must(template.New("diff").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Secret Differences</title>
	</head>
	<body>
		<div>
			<pre>
				{{- range .Diffs -}}
					{{- if eq .Type 1 -}}
						{{- .Text | printf "<ins style=\"background:#e6ffe6;\">%s</ins>" -}}
					{{- else if eq .Type -1 -}}
						{{- .Text | printf "<del style=\"background:#ffe6e6;\">%s</del>" -}}
					{{- else -}}
						{{- .Text -}}
					{{- end -}}
				{{- end -}}
			</pre>
		</div>
	</body>
	</html>
	`))
	if err := tmpl.Execute(&output, map[string]interface{}{
		"Diffs": diffs,
	}); err != nil {
		log.Fatal(err)
	}

	return output.String()
}

func confirmDiffs(secretBody, newSecretBody string) bool {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffCleanupEfficiency(dmp.DiffMain(string(secretBody), string(newSecretBody), true))

	browser.OpenReader(strings.NewReader(renderHTML(diffs)))

	log.Println(diffSummary(diffs))

	log.Println("Do you want to update the secret? [Y/n]:")
	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}
	switch strings.ToLower(string(char)) {
	case "y":
		return true
	}
	return false
}

func checkLatestSecretVersion(secret *secretsmanager.GetSecretValueOutput, awsProfile string) error {
	latestSecret, err := retrieveSecret(*secret.ARN, awsProfile)
	if err != nil {
		return err
	}

	if *latestSecret.VersionId != *secret.VersionId {
		return errors.New("newer secret version available, aborting update")
	}

	return nil
}

func diffSummary(diffs []diffmatchpatch.Diff) string {
	insertedCount := 0
	deletedCount := 0
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			insertedCount += countNewLines(diff.Text)
		case diffmatchpatch.DiffDelete:
			deletedCount += countNewLines(diff.Text)
		}
	}
	return fmt.Sprintf("Updated lines: +%d -%d (please check the changes from your browser)", insertedCount, deletedCount)
}

func countNewLines(s string) int {
	n := strings.Count(s, "\n")
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

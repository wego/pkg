package cmd

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/browser"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"github.com/wego/payments/pkg/config"
)

const (
	defaultEditor = "vim"
)

// UpdateCmd update secret on AWS Secrets Manager
func UpdateCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "update",
		Short: "Update secret on AWS Secrets Manager",
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

			secret, newSecret := openSecretInEditor(secretID, awsProfile)
			if secret == newSecret {
				log.Println("No changes in the secret, aborting...")
				os.Exit(0)
			}

			if !confirmDiffs(secret, newSecret) {
				os.Exit(0)
			}

			c := config.Load(newSecret, "", "toml")
			if err := c.Validate(); err != nil {
				log.Fatal(err)
			}

			// TODO: check existing version before update to AWS secrets manager
			log.Println("Updating secret to AWS...")
		},
	}

	return &cmd
}

func openSecretInEditor(secretID, awsProfile string) (old, new string) {
	secret, err := retrieveSecret(secretID, awsProfile)
	if err != nil {
		log.Fatal(err)
	}

	tempFileName := fmt.Sprintf("%s-%d", secretID, secret.CreatedDate.Unix())
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
		fmt.Printf("Error %s while looking up for %s!!", path, editor)
	}

	editCmdArgs := []string{}
	switch editor {
	case "code", "goland", "idea":
		editCmdArgs = append(editCmdArgs, "--wait")
	}

	f.WriteString(*secret.SecretString)
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

	return *secret.SecretString, string(newSecret)
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
	tmpl.Execute(&output, map[string]interface{}{
		"Diffs": diffs,
	})

	return output.String()
}

func confirmDiffs(secretBody, newSecretBody string) bool {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(secretBody), string(newSecretBody), true)

	browser.OpenReader(strings.NewReader(renderHTML(dmp.DiffCleanupEfficiency(diffs))))

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

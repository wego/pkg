package cmd

import (
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func retrieveSecret(secretID, awsProfile string) (*secretsmanager.GetSecretValueOutput, error) {
	ssoCmd := exec.Command("aws", "sso", "login", "--profile", awsProfile)
	if err := ssoCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := ssoCmd.Wait(); err != nil {
		log.Fatal(err)
	}

	sm := secretsmanager.New(session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config:            aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)},
			Profile:           awsProfile,
		},
	)))
	return sm.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: &secretID,
	})
}

package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/sjansen/slackbot/internal/api"
	"github.com/sjansen/slackbot/internal/app"
	"github.com/sjansen/slackbot/internal/clients"
	"github.com/sjansen/slackbot/internal/repos"
)

func main() {
	env := os.Getenv("AWS_EXECUTION_ENV")
	if env == "" {
		fmt.Fprintln(os.Stderr, "This executable is intended to run on AWS Lambda.")
		os.Exit(1)
	}

	xray.Configure(xray.Config{
		LogLevel: "info",
	})
	slack := clients.NewSlackClient(
		os.Getenv("SLACKBOT_OAUTH_ACCESS_TOKEN"),
	)

	repo, err := repos.NewWordRepo(
		os.Getenv("SLACKBOT_DYNAMODB_TABLE"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	handler := api.NewLambdaHandler(
		&app.Engine{
			Repo:  repo,
			Slack: slack,
		},
		os.Getenv("SLACKBOT_VERIFICATION_TOKEN"),
	)

	lambda.Start(handler.HandleRequest)
}

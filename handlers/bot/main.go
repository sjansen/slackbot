package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
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

	sess, err := session.NewSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	oauthAccessToken := os.Getenv("SLACKBOT_OAUTH_ACCESS_TOKEN")
	reqSigningSecret := os.Getenv("SLACKBOT_REQ_SIGNING_SECRET")
	if os.Getenv("SLACKBOT_USE_SSM") != "" {
		svc := ssm.New(sess)

		resp, err := svc.GetParameters(&ssm.GetParametersInput{
			Names: []*string{
				aws.String(oauthAccessToken),
				aws.String(reqSigningSecret),
			},
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, param := range resp.Parameters {
			switch *param.Name {
			case oauthAccessToken:
				oauthAccessToken = *param.Value
			case reqSigningSecret:
				reqSigningSecret = *param.Value
			}
		}
	}

	xray.Configure(xray.Config{
		LogLevel: "info",
	})

	slack := clients.NewSlackClient(oauthAccessToken)

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
		reqSigningSecret,
	)

	lambda.Start(handler.HandleRequest)
}

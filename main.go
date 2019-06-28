package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"

	"github.com/sjansen/slackbot/internal/repos"
)

var api *slack.Client
var table = os.Getenv("SLACKBOT_DYNAMODB_TABLE")
var verificationToken = slackevents.OptionVerifyToken(
	&slackevents.TokenComparator{
		VerificationToken: os.Getenv("SLACKBOT_VERIFICATION_TOKEN"),
	},
)

var repo *repos.WordRepo
var setWord = regexp.MustCompile(`set word (?P<word>\w+)`)

func main() {
	env := os.Getenv("AWS_EXECUTION_ENV")
	if env == "" {
		fmt.Fprintln(os.Stderr, "This executable is intended to run on AWS Lambda.")
		os.Exit(1)
	}

	xray.Configure(xray.Config{
		LogLevel: "info",
	})
	api = slack.New(
		os.Getenv("SLACKBOT_OAUTH_ACCESS_TOKEN"),
		slack.OptionHTTPClient(xray.Client(nil)),
	)

	var err error
	repo, err = repos.NewWordRepo(table)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	lambda.Start(handler)
}

func handler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("Request: %s\n", req.Body)

	event, err := slackevents.ParseEvent(
		json.RawMessage(req.Body),
		verificationToken,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		resp := &events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Cache-Control": "no-cache, no-store, must-revalidate",
				"Content-Type":  "text/text; charset=utf-8",
			},
		}
		return resp, nil
	}

	switch event.Type {
	case slackevents.CallbackEvent:
		switch ev := event.InnerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			fmt.Printf("Event Data: %#v\n", ev)
			if strings.Contains(ev.Text, "get word") {
				msg, err := repo.GetWord(ctx)
				if err != nil {
					msg = err.Error()
				} else if msg == "" {
					msg = "not set"
				}
				channelID, timestamp, err := api.PostMessageContext(
					ctx, ev.Channel,
					slack.MsgOptionText(msg, false),
				)
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
				} else {
					fmt.Printf("Message sent to %q at %s\n", channelID, timestamp)
				}
			}
		case *slackevents.MessageEvent:
			fmt.Printf("Event Data: %#v\n", ev)
			matches := setWord.FindStringSubmatch(ev.Text)
			if len(matches) > 1 {
				msg := "success"
				err := repo.SetWord(ctx, matches[1])
				if err != nil {
					msg = err.Error()
				}
				channelID, timestamp, err := api.PostMessageContext(
					ctx, ev.Channel,
					slack.MsgOptionText(msg, false),
				)
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
				} else {
					fmt.Printf("Message sent to %q at %s\n", channelID, timestamp)
				}
			}
		}
	case slackevents.URLVerification:
		cr := &slackevents.ChallengeResponse{}
		err = json.Unmarshal([]byte(req.Body), cr)
		if err != nil {
			return nil, err
		}

		resp := &events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: cr.Challenge,
		}
		return resp, nil
	}

	resp := &events.APIGatewayProxyResponse{
		StatusCode: 200,
	}
	return resp, nil
}

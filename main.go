package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

var api = slack.New(os.Getenv("SLACKBOT_OAUTH_ACCESS_TOKEN"))
var verificationToken = slackevents.OptionVerifyToken(
	&slackevents.TokenComparator{
		VerificationToken: os.Getenv("SLACKBOT_VERIFICATION_TOKEN"),
	},
)

func main() {
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
			channelID, timestamp, err := api.PostMessage(
				ev.Channel,
				slack.MsgOptionText("What's up?", false),
			)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				fmt.Printf("Message sent to %q at %s\n", channelID, timestamp)
			}
		case *slackevents.MessageEvent:
			fmt.Printf("Event Data: %#v\n", ev)
			if strings.Contains(ev.Text, "knock knock") {
				channelID, timestamp, err := api.PostMessage(
					ev.Channel,
					slack.MsgOptionText("Who's there?", false),
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

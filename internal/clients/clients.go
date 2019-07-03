package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/nlopes/slack"
)

func NewSlackClient(oauthAccessToken string) *Slack {
	api := slack.New(
		oauthAccessToken,
		slack.OptionHTTPClient(xray.Client(nil)),
	)
	return &Slack{api: api}
}

type Slack struct {
	api *slack.Client
}

func (s *Slack) PostEphemeralMessage(ctx context.Context, channel, user, text string) error {
	timestamp, err := s.api.PostEphemeralContext(
		ctx, channel, user,
		slack.MsgOptionText(text, true),
		slack.MsgOptionIconEmoji(":robot_face:"),
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Ephemeral message sent to %q@%q at %s\n", user, channel, timestamp)
	}
	return err
}

func (s *Slack) PostMessage(ctx context.Context, channel, text string) error {
	channelID, timestamp, err := s.api.PostMessageContext(
		ctx, channel,
		slack.MsgOptionText(text, true),
		slack.MsgOptionIconEmoji(":robot_face:"),
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Message sent to %q at %s\n", channelID, timestamp)
	}
	return err
}

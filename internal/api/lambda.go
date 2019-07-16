package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nlopes/slack/slackevents"
	"github.com/sjansen/slackbot/internal/app"
)

type LambdaHandler struct {
	app               *app.Engine
	verificationToken slackevents.Option
}

func NewLambdaHandler(app *app.Engine, verificationToken string) *LambdaHandler {
	return &LambdaHandler{
		app: app,
		verificationToken: slackevents.OptionVerifyToken(
			&slackevents.TokenComparator{
				VerificationToken: verificationToken,
			},
		),
	}
}

func (h *LambdaHandler) HandleRequest(ctx context.Context, req *events.APIGatewayProxyRequest) (
	*events.APIGatewayProxyResponse, error,
) {
	fmt.Printf("Request: %s\n", req.Body)

	event, err := slackevents.ParseEvent(
		json.RawMessage(req.Body),
		h.verificationToken,
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
			h.app.HandleMention(ctx, ev.Channel, ev.User, ev.Text)
		case *slackevents.MessageEvent:
			fmt.Printf("Event Data: %#v\n", ev)
			h.app.HandlePrivateMessage(ctx, ev.Channel, ev.Text)
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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"

	"github.com/sjansen/slackbot/internal/app"
)

type LambdaHandler struct {
	app           *app.Engine
	signingSecret string
}

func NewLambdaHandler(app *app.Engine, signingSecret string) *LambdaHandler {
	return &LambdaHandler{
		app:           app,
		signingSecret: signingSecret,
	}
}

func (h *LambdaHandler) HandleRequest(ctx context.Context, req *events.APIGatewayProxyRequest) (
	*events.APIGatewayProxyResponse, error,
) {
	fmt.Printf("Request: base64=%v %s\n", req.IsBase64Encoded, req.Body)

	verifier, err := slack.NewSecretsVerifier(req.MultiValueHeaders, h.signingSecret)
	if err != nil {
		return nil, err
	}
	if _, err = verifier.Write([]byte(req.Body)); err != nil {
		return nil, err
	}
	if err = verifier.Ensure(); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		resp := &events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Cache-Control": "no-cache, no-store, must-revalidate",
				"Content-Type":  "text/text; charset=utf-8",
			},
		}
		return resp, nil
	}

	event, err := slackevents.ParseEvent(
		json.RawMessage(req.Body),
		slackevents.OptionNoVerifyToken(),
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		resp := &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
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

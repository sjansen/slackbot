package app

import (
	"context"
	"regexp"
	"strings"
)

type Engine struct {
	Repo  WordRepo
	Slack SlackClient
}

type SlackClient interface {
	PostMessage(ctx context.Context, channel, text string) error
	PostEphemeralMessage(ctx context.Context, channel, user, text string) error
}

type WordRepo interface {
	GetWord(ctx context.Context) (string, error)
	SetWord(ctx context.Context, word string) error
}

var setWord = regexp.MustCompile(`set word (?P<word>\w+)`)

func (app *Engine) HandleMention(ctx context.Context, channel, user, text string) error {
	if strings.Contains(text, "get word") {
		msg, err := app.Repo.GetWord(ctx)
		if err != nil {
			msg = err.Error()
		} else if msg == "" {
			msg = "not set"
		}
		app.Slack.PostEphemeralMessage(ctx, channel, user, msg)
	}
	return nil
}

func (app *Engine) HandlePrivateMessage(ctx context.Context, channel, text string) error {
	matches := setWord.FindStringSubmatch(text)
	if len(matches) > 1 {
		msg := "success"
		err := app.Repo.SetWord(ctx, matches[1])
		if err != nil {
			msg = err.Error()
		}
		return app.Slack.PostMessage(ctx, channel, msg)
	}
	return nil
}

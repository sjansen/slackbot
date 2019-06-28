package repos

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var twoWeeks = time.Hour * 24 * 14

type secretWord struct {
	ID      string    `dynamodbav:"id,string"`
	Expires time.Time `dynamodbav:"expires,unixtime"`
	Word    string    `dynamodbav:"word,string"`
}

type WordRepo struct {
	svc   *dynamodb.DynamoDB
	table *string
}

func NewWordRepo(table string) (*WordRepo, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := dynamodb.New(sess)
	xray.AWS(svc.Client)
	repo := &WordRepo{
		svc:   svc,
		table: aws.String(table),
	}
	return repo, err
}

func (repo *WordRepo) GetWord(ctx context.Context) (string, error) {
	now := time.Now().UTC()
	id := idForDate(now)

	result, err := repo.svc.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: repo.table,
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return "", err
	}

	item := &secretWord{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return "", err
	}
	return item.Word, err
}

func (repo *WordRepo) SetWord(ctx context.Context, word string) error {
	now := time.Now().UTC()
	id := idForDate(now)
	expires := now.Add(twoWeeks)

	av, err := dynamodbattribute.MarshalMap(&secretWord{
		ID:      id,
		Expires: expires,
		Word:    word,
	})
	if err != nil {
		return err
	}

	_, err = repo.svc.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      av,
		TableName: repo.table,
	})
	return err
}

func idForDate(date time.Time) string {
	return date.Format("word:2006-01-02")
}

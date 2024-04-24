package board

import (
	"encoding/json"
	"os"

	log "github.com/Ptt-Alertor/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/myutil"
)

const tableName string = "boards"

// column: Board, Articles
type DynamoDB struct {
}

func (DynamoDB) GetArticles(boardName string) (articles article.Articles) {
	if boardName == "" {
		return
	}

	// dynamo := dynamodb.New(session.New())
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Endpoint:    aws.String(os.Getenv("DB_CONNECTION")),
		Credentials: credentials.NewStaticCredentials("local", "local", ""),
		// CredentialsChainVerboseErrors: aws.Bool(false),
	})
	dynamo := dynamodb.New(sess)
	result, err := dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Board": {
				S: aws.String(boardName),
			},
		},
	})
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DynamoDB Find Board Failed")
		return
	}

	if len(result.Item) == 0 {
		log.WithField("board", boardName).Warn("Board Not Found")
		return
	}

	articlesJSON := aws.StringValue(result.Item["Articles"].S)

	if articlesJSON != "" {
		err = json.Unmarshal([]byte(articlesJSON), &articles)
		if err != nil {
			myutil.LogJSONDecode(err, articlesJSON)
		}
	}
	return articles
}

func (DynamoDB) Save(boardName string, articles article.Articles) error {
	articlesJSON, err := json.Marshal(articles)
	if err != nil {
		myutil.LogJSONEncode(err, articles)
		return err
	}

	// dynamo := dynamodb.New(session.New())
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Endpoint:    aws.String(os.Getenv("DB_CONNECTION")),
		Credentials: credentials.NewStaticCredentials("local", "local", ""),
		// CredentialsChainVerboseErrors: aws.Bool(false),
	})
	dynamo := dynamodb.New(sess)
	_, err = dynamo.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Board": {
				S: aws.String(boardName),
			},
			"Articles": {
				S: aws.String(string(articlesJSON)),
			},
		},
		TableName: aws.String(tableName),
	})

	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DynamoDB Save Board Failed")
	}
	return err
}

func (DynamoDB) Delete(boardName string) error {
	// dynamo := dynamodb.New(session.New())
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Endpoint:    aws.String(os.Getenv("DB_CONNECTION")),
		Credentials: credentials.NewStaticCredentials("local", "local", ""),
		// CredentialsChainVerboseErrors: aws.Bool(false),
	})
	dynamo := dynamodb.New(sess)
	_, err := dynamo.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Board": {
				S: aws.String(boardName),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.WithField("runtime", myutil.BasicRuntimeInfo()).WithError(err).Error("DynamoDB Delete Board Failed")
	}

	return err
}

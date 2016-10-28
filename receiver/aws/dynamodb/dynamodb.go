package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoDBClient struct {
	client *dynamodb.DynamoDB
}

// NewDynamoDBClient creates new DynamoDBClient object
func NewDynamoDBClient() *DynamoDBClient {
	return &DynamoDBClient{
		client: dynamodb.New(session.New(), &aws.Config{}),
	}
}

// List returns all items in the given table
func (c *DynamoDBClient) List(table string) ([]map[string]*dynamodb.AttributeValue, error) {
	resp, err := c.client.Scan(&dynamodb.ScanInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

// Select returns matched items in the given table
func (c *DynamoDBClient) Select(table, index string, filter map[string]string) ([]map[string]*dynamodb.AttributeValue, error) {
	keyConditions := make(map[string]*dynamodb.Condition)

	for k, v := range filter {
		keyConditions[k] = &dynamodb.Condition{
			ComparisonOperator: aws.String(dynamodb.ComparisonOperatorEq),
			AttributeValueList: []*dynamodb.AttributeValue{
				&dynamodb.AttributeValue{
					S: aws.String(v),
				},
			},
		}
	}

	params := &dynamodb.QueryInput{
		TableName:     aws.String(table),
		KeyConditions: keyConditions,
	}

	if index != "" {
		params.IndexName = aws.String(index)
	}

	resp, err := c.client.Query(params)
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

// Create create new item in the given table
func (c *DynamoDBClient) Create(table string, fields map[string]string) error {
	item := make(map[string]*dynamodb.AttributeValue)

	for k, v := range fields {
		item[k] = &dynamodb.AttributeValue{
			S: aws.String(v),
		}
	}

	_, err := c.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

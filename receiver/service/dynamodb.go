package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	dynamodbSvc = dynamodb.New(session.New(), &aws.Config{})
)

// List returns all items in the given table
func List(table string) ([]map[string]*dynamodb.AttributeValue, error) {
	resp, err := dynamodbSvc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

// Select returns matched items in the given table
func Select(table, index string, filter map[string]string) ([]map[string]*dynamodb.AttributeValue, error) {
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

	resp, err := dynamodbSvc.Query(params)
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

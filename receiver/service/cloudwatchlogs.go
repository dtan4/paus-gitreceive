package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var (
	cloudwatchlogsSvc = cloudwatchlogs.New(session.New(), &aws.Config{})
)

// CreateLogGroup creates new CloudWatch Logs log group
func CreateLogGroup(name string) error {
	_, err := cloudwatchlogsSvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})

	if err != nil {
		return err
	}

	return nil
}

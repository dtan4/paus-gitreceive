package cloudwatchlogs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type CloudWatchLogsClient struct {
	client *cloudwatchlogs.CloudWatchLogs
}

// NewCloudWatchLogsClient creates new CloudWatchLogsClient object
func NewCloudWatchLogsClient() *CloudWatchLogsClient {
	return &CloudWatchLogsClient{
		client: cloudwatchlogs.New(session.New(), &aws.Config{}),
	}
}

// CreateLogGroup creates new CloudWatch Logs log group
func (c *CloudWatchLogsClient) CreateLogGroup(name string) error {
	_, err := c.client.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})

	if err != nil {
		return err
	}

	return nil
}

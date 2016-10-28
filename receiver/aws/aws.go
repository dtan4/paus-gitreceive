package aws

import (
	"github.com/dtan4/paus-gitreceive/receiver/aws/cloudwatchlogs"
	"github.com/dtan4/paus-gitreceive/receiver/aws/dynamodb"
	"github.com/dtan4/paus-gitreceive/receiver/aws/ec2"
	"github.com/dtan4/paus-gitreceive/receiver/aws/ecr"
	"github.com/dtan4/paus-gitreceive/receiver/aws/ecs"
	"github.com/dtan4/paus-gitreceive/receiver/aws/sts"
)

var (
	cloudWatchLogsClient *cloudwatchlogs.CloudWatchLogsClient
	dynamoDBClient       *dynamodb.DynamoDBClient
	ec2Client            *ec2.EC2Client
	ecsClient            *ecs.ECSClient
	ecrClient            *ecr.ECRClient
	stsClient            *sts.STSClient
)

// CloudWatchLogs returns CloudWatchLogs object and create new one if it does not exist
func CloudWatchLogs() *cloudwatchlogs.CloudWatchLogsClient {
	if cloudWatchLogsClient == nil {
		cloudWatchLogsClient = cloudwatchlogs.NewCloudWatchLogsClient()
	}

	return cloudWatchLogsClient
}

// DynamoDB returns DynamoDB object and create new one if it does not exist
func DynamoDB() *dynamodb.DynamoDBClient {
	if dynamoDBClient == nil {
		dynamoDBClient = dynamodb.NewDynamoDBClient()
	}

	return dynamoDBClient
}

// EC2 returns EC2 object and create new one if it does not exist
func EC2() *ec2.EC2Client {
	if ec2Client == nil {
		ec2Client = ec2.NewEC2Client()
	}

	return ec2Client
}

// ECS returns ECS object and create new one if it does not exist
func ECS() *ecs.ECSClient {
	if ecsClient == nil {
		ecsClient = ecs.NewECSClient()
	}

	return ecsClient
}

// ECR returns ECR object and create new one if it does not exist
func ECR() *ecr.ECRClient {
	if ecrClient == nil {
		ecrClient = ecr.NewECRClient()
	}

	return ecrClient
}

// STS return STS object and create new one if it does not exist
func STS() *sts.STSClient {
	if stsClient == nil {
		stsClient = sts.NewSTSClient()
	}

	return stsClient
}

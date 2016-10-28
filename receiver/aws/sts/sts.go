package sts

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type STSClient struct {
	client *sts.STS
}

// NewSTSClient creates new STSClient object
func NewSTSClient() *STSClient {
	return &STSClient{
		client: sts.New(session.New(), &aws.Config{}),
	}
}

// GetAWSAccountID returns AWS account ID of current log in account
func (c *STSClient) GetAWSAccountID() (string, error) {
	resp, err := c.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *resp.Account, nil
}

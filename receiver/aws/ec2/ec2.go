package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Client struct {
	client *ec2.EC2
}

// NewEC2Client creates new EC2Client object
func NewEC2Client() *EC2Client {
	return &EC2Client{
		client: ec2.New(session.New(), &aws.Config{}),
	}
}

// GetInstance returns instance with the given ID
func (c *EC2Client) GetInstance(instanceID string) (*ec2.Instance, error) {
	resp, err := c.client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})

	if err != nil {
		return nil, err
	}

	return resp.Reservations[0].Instances[0], nil
}

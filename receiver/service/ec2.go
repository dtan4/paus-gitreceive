package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	ec2Svc = ec2.New(session.New(), &aws.Config{})
)

// GetInstance returns instance with the given ID
func GetInstance(instanceID string) (*ec2.Instance, error) {
	resp, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})

	if err != nil {
		return nil, err
	}

	return resp.Reservations[0].Instances[0], nil
}

package aws

import (
	"github.com/dtan4/paus-gitreceive/receiver/aws/ec2"
	"github.com/dtan4/paus-gitreceive/receiver/aws/ecr"
	"github.com/dtan4/paus-gitreceive/receiver/aws/sts"
)

var (
	ec2Client *ec2.EC2Client
	ecrClient *ecr.ECRClient
	stsClient *sts.STSClient
)

// EC2 returns EC2 object and create new one if it does not exist
func EC2() *ec2.EC2Client {
	if ec2Client == nil {
		ec2Client = ec2.NewEC2Client()
	}

	return ec2Client
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

package aws

import (
	"github.com/dtan4/paus-gitreceive/receiver/aws/ecr"
	"github.com/dtan4/paus-gitreceive/receiver/aws/sts"
)

var (
	ecrClient *ecr.ECRClient
	stsClient *sts.STSClient
)

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

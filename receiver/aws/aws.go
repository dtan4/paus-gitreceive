package aws

import (
	"github.com/dtan4/paus-gitreceive/receiver/aws/ecr"
)

var (
	ecrClient *ecr.ECRClient
)

// ECR returns ECR object and create new one if it does not exist
func ECR() *ecr.ECRClient {
	if ecrClient == nil {
		ecrClient = ecr.NewECRClient()
	}

	return ecrClient
}

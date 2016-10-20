package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func GetECRToken(registryID string) (string, error) {
	svc := ecr.New(session.New(), &aws.Config{})

	resp, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{
			aws.String(registryID),
		},
	})
	if err != nil {
		return "", err
	}

	return *resp.AuthorizationData[0].AuthorizationToken, nil
}

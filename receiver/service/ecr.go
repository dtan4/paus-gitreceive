package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func CreateRepository(registryID, repository string) error {
	svc := ecr.New(session.New(), &aws.Config{})

	_, err := svc.CreateRepository(&ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repository),
	})

	if err != nil {
		return err
	}

	return nil
}

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

func RepositoryExists(registryID, repository string) (bool, error) {
	svc := ecr.New(session.New(), &aws.Config{})

	resp, err := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{
			aws.String(repository),
		},
	})

	if err != nil {
		return false, err
	}

	if len(resp.Repositories) == 0 {
		return false, nil
	}

	return true, nil
}

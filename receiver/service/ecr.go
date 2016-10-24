package service

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/fsouza/go-dockerclient"
)

// CreateRepository creates new Repository
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

// GetECRAuthConf returns ECR authrization configuration
func GetECRAuthConf(registryID string) (docker.AuthConfiguration, error) {
	svc := ecr.New(session.New(), &aws.Config{})

	resp, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{
			aws.String(registryID),
		},
	})
	if err != nil {
		return docker.AuthConfiguration{}, err
	}

	username, token, err := extractToken(resp.AuthorizationData[0])
	if err != nil {
		return docker.AuthConfiguration{}, err
	}

	return docker.AuthConfiguration{
		Username:      username,
		Password:      token,
		Email:         "",
		ServerAddress: "",
	}, nil
}

// GetRegistryDomain returns fully-qualified ECR registry domain
func GetRegistryDomain(accountID, region string) string {
	return accountID + ".dkr.ecr." + region + ".amazonaws.com"
}

// RepositoryExists returns whether the specified repository exists or not
func RepositoryExists(registryID, repository string) bool {
	svc := ecr.New(session.New(), &aws.Config{})

	resp, err := svc.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{
			aws.String(repository),
		},
	})

	if err != nil {
		return false
	}

	return len(resp.Repositories) > 0
}

func extractToken(authData *ecr.AuthorizationData) (string, string, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(aws.StringValue(authData.AuthorizationToken))
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(string(decodedToken), ":", 2)

	return parts[0], parts[1], nil
}

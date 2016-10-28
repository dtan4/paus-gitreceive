package ecr

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/fsouza/go-dockerclient"
)

type ECRClient struct {
	client *ecr.ECR
}

// NewECR creates new ECR object
func NewECRClient() *ECRClient {
	return &ECRClient{
		client: ecr.New(session.New(), &aws.Config{}),
	}
}

// CreateRepository creates new Repository
func (c *ECRClient) CreateRepository(registryID, repository string) error {
	_, err := c.client.CreateRepository(&ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repository),
	})

	if err != nil {
		return err
	}

	return nil
}

// GetECRAuthConf returns ECR authrization configuration
func (c *ECRClient) GetECRAuthConf(registryID string) (docker.AuthConfiguration, error) {
	resp, err := c.client.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
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
func (c *ECRClient) GetRegistryDomain(accountID, region string) string {
	return accountID + ".dkr.ecr." + region + ".amazonaws.com"
}

// RepositoryExists returns whether the specified repository exists or not
func (c *ECRClient) RepositoryExists(registryID, repository string) bool {
	resp, err := c.client.DescribeRepositories(&ecr.DescribeRepositoriesInput{
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

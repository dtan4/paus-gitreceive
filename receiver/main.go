package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	DefaultDockerHost    = "tcp://localhost:2375"
	DefaultEtcdEndpoint  = "http://localhost:2379"
	DefaultRepositoryDir = "/repos"
)

func unpackReceivedFiles(repositoryDir, username, projectName string, stdin io.Reader) (string, error) {
	repositoryPath := filepath.Join(repositoryDir, username, projectName)

	if err := os.MkdirAll(repositoryPath, 0777); err != nil {
		return "", err
	}

	reader := tar.NewReader(stdin)

	for {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		buffer := new(bytes.Buffer)
		outPath := filepath.Join(repositoryPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err = os.Stat(outPath); err != nil {
				os.MkdirAll(outPath, 0755)
			}

		case tar.TypeReg, tar.TypeRegA:
			if _, err = io.Copy(buffer, reader); err != nil {
				return "", err
			}

			if err = ioutil.WriteFile(outPath, buffer.Bytes(), os.FileMode(header.Mode)); err != nil {
				return "", err
			}
		}
	}

	return repositoryPath, nil
}

func main() {
	baseWorkingDir, _ := os.Getwd()

	dockerHost := os.Getenv("DOCKER_HOST")

	if dockerHost == "" {
		dockerHost = DefaultDockerHost
	}

	etcdEndpoint := os.Getenv("ETCD_ENDPOINT")

	if etcdEndpoint == "" {
		etcdEndpoint = DefaultEtcdEndpoint
	}

	repositoryDir := os.Getenv("REPOSITORY_DIR")

	if repositoryDir == "" {
		repositoryDir = DefaultRepositoryDir
	}

	commitMetadata := NewCommitMetadataFromArgs(os.Args[1:])

	fmt.Println(baseWorkingDir)
	fmt.Println(commitMetadata.Repository)
	fmt.Println(commitMetadata.Revision)
	fmt.Println(commitMetadata.Username)
	fmt.Println(commitMetadata.AppName)
	fmt.Println(commitMetadata.ProjectName)

	repositoryPath, err := unpackReceivedFiles(repositoryDir, commitMetadata.Username, commitMetadata.AppName, os.Stdin)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(repositoryPath)

	if err = os.Chdir(repositoryPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, "=====> docker-compose.yml was NOT found!")
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")
	compose := NewCompose(composeFilePath, commitMetadata.ProjectName)

	fmt.Println("=====> Building ...")

	if err = compose.Build(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Pulling ...")

	if err = compose.Pull(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Deploying ...")

	if err = compose.Up(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	webContainerId, err := compose.GetContainerId("web")

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(webContainerId)
}

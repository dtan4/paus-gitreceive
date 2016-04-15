package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DefaultDockerHost    = "tcp://localhost:2375"
	DefaultEtcdEndpoint  = "http://localhost:2379"
	DefaultRepositoryDir = "/repos"
)

func runCommand(command *exec.Cmd) error {
	stdout, err := command.StdoutPipe()

	if err != nil {
		return err
	}

	command.Start()
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	command.Wait()

	return nil
}

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

	if _, err = os.Stat("docker-compose.yml"); err != nil {
		fmt.Fprintln(os.Stderr, "=====> docker-compose.yml was NOT found!")
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")

	fmt.Println("=====> Building ...")
	cmd := exec.Command("docker-compose", "-p", commitMetadata.ProjectName, "build")

	if err = runCommand(cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Pulling ...")
	cmd = exec.Command("docker-compose", "-p", commitMetadata.ProjectName, "pull")

	if err = runCommand(cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Deploying ...")
	cmd = exec.Command("docker-compose", "-p", commitMetadata.ProjectName, "up", "-d")

	if err = runCommand(cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

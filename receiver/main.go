package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DefaultDockerHost    = "tcp://localhost:2375"
	DefaultEtcdEndpoint  = "http://localhost:2379"
	DefaultRepositoryDir = "/repos"
)

func deploy(commitMetadata *CommitMetadata, composeFilePath string) (string, error) {
	var err error

	compose := NewCompose(composeFilePath, commitMetadata.ProjectName)

	fmt.Println("=====> Building ...")

	if err = compose.Build(); err != nil {
		return "", err
	}

	fmt.Println("=====> Pulling ...")

	if err = compose.Pull(); err != nil {
		return "", err
	}

	fmt.Println("=====> Deploying ...")

	if err = compose.Up(); err != nil {
		return "", err
	}

	webContainerId, err := compose.GetContainerId("web")

	if err != nil {
		return "", err
	}

	return webContainerId, nil
}

func registerApplicationMetadata(commitMetadata *CommitMetadata, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + commitMetadata.Username

	if etcd.HasKey(userDirectoryKey) {
		_ = etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/" + commitMetadata.AppName

	if etcd.HasKey(appDirectoryKey) {
		_ = etcd.Mkdir(appDirectoryKey)
		_ = etcd.Mkdir(appDirectoryKey + "/envs")
		_ = etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := etcd.Set(appDirectoryKey+"/revisions/"+commitMetadata.Revision, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return err
	}

	return nil
}

func registerVulcandInformation(commitMetadata *CommitMetadata, baseDomain string, webContainer *Container, etcd *Etcd) error {
	vulcandDirectoryKeyBase := "/vulcand"

	// {"Type": "http"}
	if err := etcd.Set(vulcandDirectoryKeyBase+"/backends/"+commitMetadata.ProjectName+"/backend", "{\"Type\": \"http\"}"); err != nil {
		return err
	}

	// {"URL": "http://$web_container_host_ip:$web_container_port"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+commitMetadata.ProjectName+"/servers/"+webContainer.ContainerId,
		"{\"URL\": \"http://"+webContainer.HostIP()+":"+webContainer.HostPort()+"\"}",
	); err != nil {
		return err
	}

	// {"URL": "http://$web_container_host_ip:$web_container_port"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+commitMetadata.ProjectName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \"$PROJECT_NAME\", \"Route\": \"Host(`"+commitMetadata.ProjectName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+commitMetadata.Username+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \"$PROJECT_NAME\", \"Route\": \"Host(`"+commitMetadata.Username+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+commitMetadata.AppName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \"$PROJECT_NAME\", \"Route\": \"Host(`"+commitMetadata.AppName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

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
	baseDomain := "pausapp.com"

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
	webContainerId, err := deploy(commitMetadata, composeFilePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	webContainer, err := NewContainer(dockerHost, webContainerId)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	etcd, err := NewEtcd(etcdEndpoint)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = registerApplicationMetadata(commitMetadata, etcd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = registerVulcandInformation(commitMetadata, baseDomain, webContainer, etcd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
	"strings"
	"time"
)

func deploy(dockerHost string, application *Application, composeFilePath string) (string, error) {
	var err error

	compose := NewCompose(dockerHost, composeFilePath, application.ProjectName)

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

func injectEnvironmentVariables(application *Application, composeFile *ComposeFile, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil
	}

	appDirectoryKey := userDirectoryKey + "/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !etcd.HasKey(envDirectoryKey) {
		return nil
	}

	envKeys, err := etcd.List(envDirectoryKey, false)

	if err != nil {
		return err
	}

	environmentVariables := map[string]string{}

	for _, key := range envKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return err
		}

		environmentVariables[strings.Replace(key, envDirectoryKey, "", 1)] = value
	}

	composeFile.InjectEnvironmentVariables(environmentVariables)

	return nil
}

func registerApplicationMetadata(application *Application, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if etcd.HasKey(userDirectoryKey) {
		_ = etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/" + application.AppName

	if etcd.HasKey(appDirectoryKey) {
		_ = etcd.Mkdir(appDirectoryKey)
		_ = etcd.Mkdir(appDirectoryKey + "/envs")
		_ = etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := etcd.Set(appDirectoryKey+"/revisions/"+application.Revision, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return err
	}

	return nil
}

func registerVulcandInformation(application *Application, baseDomain string, webContainer *Container, etcd *Etcd) error {
	vulcandDirectoryKeyBase := "/vulcand"

	// {"Type": "http"}
	if err := etcd.Set(vulcandDirectoryKeyBase+"/backends/"+application.ProjectName+"/backend", "{\"Type\": \"http\"}"); err != nil {
		return err
	}

	// {"URL": "http://$web_container_host_ip:$web_container_port"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+application.ProjectName+"/servers/"+webContainer.ContainerId,
		"{\"URL\": \"http://"+webContainer.HostIP()+":"+webContainer.HostPort()+"\"}",
	); err != nil {
		return err
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$PROJECT_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.ProjectName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.ProjectName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$USER_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.Username+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.Username+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$APP_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.AppName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.AppName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return err
	}

	return nil
}

func removeUnpackedFiles(repositoryPath, newComposeFilePath string) error {
	files, err := ioutil.ReadDir(repositoryPath)

	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Join(repositoryPath, file.Name()) != newComposeFilePath {
			if err = os.RemoveAll(filepath.Join(repositoryPath, file.Name())); err != nil {
				return err
			}
		}
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
	config, err := LoadConfig()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	application := ApplicationFromArgs(os.Args[1:])
	repositoryPath, err := unpackReceivedFiles(config.RepositoryDir, application.Username, application.ProjectName, os.Stdin)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, "=====> docker-compose.yml was NOT found!")
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")
	composeFile, err := NewComposeFile(composeFilePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	etcd, err := NewEtcd(config.EtcdEndpoint)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = injectEnvironmentVariables(application, composeFile, etcd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	newComposeFilePath := filepath.Join(repositoryPath, "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10)+".yml")

	if err = composeFile.SaveAs(newComposeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	webContainerId, err := deploy(config.DockerHost, application, newComposeFilePath)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Application container is launched.")

	webContainer, err := ContainerFromID(config.DockerHost, webContainerId)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Registering metadata ...")

	if err = registerApplicationMetadata(application, etcd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = registerVulcandInformation(application, config.BaseDomain, webContainer, etcd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	urlList := []string{
		"http://" + application.ProjectName + "." + config.BaseDomain,
		"http://" + application.Username + "." + config.BaseDomain,
		"http://" + application.AppName + "." + config.BaseDomain,
	}

	fmt.Println("=====> " + application.Repository + " was successfully deployed at:")

	for _, url := range urlList {
		fmt.Println("         " + url)
	}

	if err = removeUnpackedFiles(repositoryPath, newComposeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

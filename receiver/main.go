package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func appDirExists(application *Application, etcd *Etcd) bool {
	return etcd.HasKey("/paus/users/" + application.Username + "/apps/" + application.AppName)
}

func deploy(dockerHost string, application *Application, composeFilePath string) (string, error) {
	var err error

	compose := NewCompose(dockerHost, composeFilePath, application.ProjectName)

	fmt.Println("=====> Building ...")

	if err = compose.Build(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to build application image. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	fmt.Println("=====> Pulling ...")

	if err = compose.Pull(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to pull application image. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	fmt.Println("=====> Deploying ...")

	if err = compose.Up(); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to start application. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	webContainerId, err := compose.GetContainerId("web")

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to web container ID. appName: %s, composeFilePath: %s", application.AppName, composeFilePath))
	}

	return webContainerId, nil
}

func getSubmodules(repositoryPath string) error {
	dir := filepath.Join(repositoryPath, ".git")

	stat, err := os.Stat(dir)

	if err == nil && stat.IsDir() {
		if e := os.RemoveAll(dir); e != nil {
			return errors.Wrap(e, fmt.Sprintf("Failed to remove %s.", dir))
		}
	}

	cmd := exec.Command("/usr/local/bin/get-submodules")

	if err = RunCommand(cmd); err != nil {
		return errors.Wrap(err, "Failed to get submodules.")
	}

	return nil
}

func injectBuildArgs(application *Application, composeFile *ComposeFile, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil
	}

	buildArgsKey := appDirectoryKey + "/build-args/"

	if !etcd.HasKey(buildArgsKey) {
		return nil
	}

	buildArgKeys, err := etcd.List(buildArgsKey, false)

	if err != nil {
		return errors.Wrap(err, "Failed to get build arg keys.")
	}

	buildArgs := map[string]string{}

	for _, key := range buildArgKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to get build arg value. key: %s", key))
		}

		buildArgs[strings.Replace(key, buildArgsKey, "", 1)] = value
	}

	composeFile.InjectBuildArgs(buildArgs)

	return nil
}

func injectEnvironmentVariables(application *Application, composeFile *ComposeFile, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		return nil
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		return nil
	}

	envDirectoryKey := appDirectoryKey + "/envs/"

	if !etcd.HasKey(envDirectoryKey) {
		return nil
	}

	envKeys, err := etcd.List(envDirectoryKey, false)

	if err != nil {
		return errors.Wrap(err, "Failed to get environment variable keys.")
	}

	environmentVariables := map[string]string{}

	for _, key := range envKeys {
		value, err := etcd.Get(key)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to get environment variable value. key: %s", key))
		}

		environmentVariables[strings.Replace(key, envDirectoryKey, "", 1)] = value
	}

	composeFile.InjectEnvironmentVariables(environmentVariables)

	return nil
}

func registerApplicationMetadata(application *Application, etcd *Etcd) error {
	userDirectoryKey := "/paus/users/" + application.Username

	if !etcd.HasKey(userDirectoryKey) {
		_ = etcd.Mkdir(userDirectoryKey)
	}

	appDirectoryKey := userDirectoryKey + "/apps/" + application.AppName

	if !etcd.HasKey(appDirectoryKey) {
		_ = etcd.Mkdir(appDirectoryKey)
		_ = etcd.Mkdir(appDirectoryKey + "/envs")
		_ = etcd.Mkdir(appDirectoryKey + "/revisions")
	}

	if err := etcd.Set(appDirectoryKey+"/revisions/"+application.Revision, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		return errors.Wrap(err, "Failed to set revisdion.")
	}

	return nil
}

func registerVulcandInformation(application *Application, baseDomain string, webContainer *Container, etcd *Etcd) error {
	vulcandDirectoryKeyBase := "/vulcand"

	// {"Type": "http"}
	if err := etcd.Set(vulcandDirectoryKeyBase+"/backends/"+application.ProjectName+"/backend", "{\"Type\": \"http\"}"); err != nil {
		return errors.Wrap(err, "Failed to set vulcand backend.")
	}

	// {"URL": "http://$web_container_host_ip:$web_container_port"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/backends/"+application.ProjectName+"/servers/"+webContainer.ContainerId,
		"{\"URL\": \"http://"+webContainer.HostIP()+":"+webContainer.HostPort()+"\"}",
	); err != nil {
		return errors.Wrap(err, "Failed to set vulcand backend server.")
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$PROJECT_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.ProjectName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.ProjectName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return errors.Wrap(err, "Failed to set vulcand frontend with project name.")
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$USER_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.Username+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.Username+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return errors.Wrap(err, "Failed to set vulcand frontend with username.")
	}

	// {"Type": "http", "BackendId": "$PROJECT_NAME", "Route": "Host(`$APP_NAME.$BASE_DOMAIN`) && PathRegexp(`/`)"}
	if err := etcd.Set(
		vulcandDirectoryKeyBase+"/frontends/"+application.AppName+"/frontend",
		"{\"Type\": \"http\", \"BackendId\": \""+application.ProjectName+"\", \"Route\": \"Host(`"+application.AppName+"."+baseDomain+"`) && PathRegexp(`/`)\"}",
	); err != nil {
		return errors.Wrap(err, "Failed to set vulcand frontend with appName.")
	}

	return nil
}

func removeUnpackedFiles(repositoryPath, newComposeFilePath string) error {
	files, err := ioutil.ReadDir(repositoryPath)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to open %s.", repositoryPath))
	}

	for _, file := range files {
		if filepath.Join(repositoryPath, file.Name()) != newComposeFilePath {
			path := filepath.Join(repositoryPath, file.Name())

			if err = os.RemoveAll(path); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to remove files in %s.", path))
			}
		}
	}

	return nil
}

func unpackReceivedFiles(repositoryDir, username, projectName string, stdin io.Reader) (string, error) {
	repositoryPath := filepath.Join(repositoryDir, username, projectName)

	if err := os.MkdirAll(repositoryPath, 0777); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to create directory %s.", repositoryPath))
	}

	reader := tar.NewReader(stdin)

	for {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return "", errors.Wrap(err, "Failed to iterate tarball.")
		}

		buffer := new(bytes.Buffer)
		outPath := filepath.Join(repositoryPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err = os.Stat(outPath); err != nil {
				if err = os.MkdirAll(outPath, 0755); err != nil {
					return "", errors.Wrap(err, fmt.Sprintf("Failed to create directory %s from tarball.", outPath))
				}
			}

		case tar.TypeReg, tar.TypeRegA:
			if _, err = io.Copy(buffer, reader); err != nil {
				return "", errors.Wrap(err, fmt.Sprintf("Failed to copy file contents in %s from tarball.", outPath))
			}

			if err = ioutil.WriteFile(outPath, buffer.Bytes(), os.FileMode(header.Mode)); err != nil {
				return "", errors.Wrap(err, fmt.Sprintf("Failed to create file %s from tarball.", outPath))
			}
		}
	}

	return repositoryPath, nil
}

func main() {
	config, err := LoadConfig()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	etcd, err := NewEtcd(config.EtcdEndpoint)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	application := ApplicationFromArgs(os.Args[1:])

	if !appDirExists(application, etcd) {
		fmt.Fprintln(os.Stderr, "=====> Application not found: "+application.AppName)
		os.Exit(1)
	}

	repositoryPath, err := unpackReceivedFiles(config.RepositoryDir, application.Username, application.ProjectName, os.Stdin)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = os.Chdir(repositoryPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println("=====> Getting submodules ...")

	if err = getSubmodules(repositoryPath); err != nil {
		errors.Fprint(os.Stderr, errors.Wrap(err, fmt.Sprintf("Failed to get submodules. path: %s", repositoryPath)))
		os.Exit(1)
	}

	composeFilePath := filepath.Join(repositoryPath, "docker-compose.yml")

	if _, err := os.Stat(composeFilePath); err != nil {
		fmt.Fprintln(os.Stderr, "=====> docker-compose.yml was NOT found!")
		os.Exit(1)
	}

	fmt.Println("=====> docker-compose.yml was found")
	composeFile, err := NewComposeFile(composeFilePath)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = injectBuildArgs(application, composeFile, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = injectEnvironmentVariables(application, composeFile, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	composeFile.RewritePortBindings()
	newComposeFilePath := filepath.Join(repositoryPath, "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10)+".yml")

	if err = composeFile.SaveAs(newComposeFilePath); err != nil {
		errors.Fprint(os.Stderr, err)
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
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("=====> Registering metadata ...")

	if err = registerApplicationMetadata(application, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = registerVulcandInformation(application, config.BaseDomain, webContainer, etcd); err != nil {
		errors.Fprint(os.Stderr, err)
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

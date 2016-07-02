package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/libcompose/config"
)

const (
	dockerHost  = "unix:///var/run/docker.sock"
	projectName = "paustest"
)

var (
	v1FilePath, v2FilePath, v2FilePathBuildArg, v2FilePathNoBuildEnv string
	v1Compose, v2Compose, v2ComposeBuildArg, v2ComposeNoBuildEnv     *Compose
)

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))

	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]

	return ok
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

func fixturePath(name string) string {
	workingDir, _ := os.Getwd()

	return filepath.Join(workingDir, "..", "fixtures", name)
}

func setup() {
	v1FilePath = fixturePath("docker-compose-v1.yml")
	v2FilePath = fixturePath("docker-compose-v2.yml")
	v2FilePathBuildArg = fixturePath("docker-compose-v2-buildarg.yml")
	v2FilePathNoBuildEnv = fixturePath("docker-compose-v2-nobuildenv.yml")

	v1Compose, _ = NewCompose(dockerHost, v1FilePath, projectName)
	v2Compose, _ = NewCompose(dockerHost, v2FilePath, projectName)
	v2ComposeBuildArg, _ = NewCompose(dockerHost, v2FilePathBuildArg, projectName)
	v2ComposeNoBuildEnv, _ = NewCompose(dockerHost, v2FilePathNoBuildEnv, projectName)
}

func TestInjectBuildArgs(t *testing.T) {
	var (
		buildArgs map[string]string
	)

	setup()

	buildArgs = map[string]string{
		"FOO": "hoge",
		"BAR": "fuga",
		"BAZ": "piyo",
	}

	// TODO: test for V1

	v2Compose.InjectBuildArgs(buildArgs)
	svc, _ := v2Compose.project.ServiceConfigs.Get("web")

	for key, _ := range buildArgs {
		if _, ok := svc.Build.Args[key]; !ok {
			t.Fatalf("Compose File V2 does not contain %s", key)
		}
	}

	buildArgs = map[string]string{
		"FOO": "hogefugapiyo",
	}
	v2ComposeBuildArg.InjectBuildArgs(buildArgs)
	svc, _ = v2ComposeBuildArg.project.ServiceConfigs.Get("web")

	oldBuildArg := "hoge"
	newBuildArg := "hogefugapiyo"

	if svc.Build.Args["FOO"] == oldBuildArg {
		t.Fatalf("Failed to update existing key FOO. FOO=%s still exists.", oldBuildArg)
	}

	if svc.Build.Args["FOO"] != newBuildArg {
		t.Fatalf("Failed to update existing key FOO. FOO=%s does not exist.", newBuildArg)
	}

	v2ComposeNoBuildEnv.InjectBuildArgs(buildArgs)
	svc, _ = v2ComposeNoBuildEnv.project.ServiceConfigs.Get("web")

	if len(svc.Build.Args) == 0 {
		t.Fatalf("Build section was created in Compose file without build section.")
	}
}

func TestInjectEnvironmentVariables(t *testing.T) {
	var (
		environmentVariables map[string]string
		envString            string
		svc                  *config.ServiceConfig
		webEnvironment       []string
	)

	setup()

	environmentVariables = map[string]string{
		"FOO": "hoge",
		"BAR": "fuga",
		"BAZ": "piyo",
	}

	v1Compose.InjectEnvironmentVariables(environmentVariables)
	svc, _ = v1Compose.project.ServiceConfigs.Get("web")
	webEnvironment = svc.Environment

	for key, value := range environmentVariables {
		envString = fmt.Sprintf("%s=%s", key, value)

		if !contains(webEnvironment, envString) {
			t.Fatalf("Compose File V1 does not contain %s", key)
		}
	}

	if !contains(webEnvironment, "DATABASE_HOST=db") {
		t.Fatalf("Original string DATABASE_HOST=db is dismissed. environments: %v", webEnvironment)
	}

	v2Compose.InjectEnvironmentVariables(environmentVariables)
	svc, _ = v2Compose.project.ServiceConfigs.Get("web")
	webEnvironment = svc.Environment

	for key, value := range environmentVariables {
		envString = fmt.Sprintf("%s=%s", key, value)

		if !contains(webEnvironment, envString) {
			t.Fatalf("Compose File V2 does not contain %s", key)
		}
	}

	environmentVariables = map[string]string{
		"FOO": "hogefugapiyo",
	}

	v2Compose.InjectEnvironmentVariables(environmentVariables)
	webEnvironment = svc.Environment

	oldEnvString := "FOO=hoge"
	newEnvString := "FOO=hogefugapiyo"

	if contains(webEnvironment, oldEnvString) {
		t.Fatalf("Failed to update existing key FOO. %s still exists.", oldEnvString)
	}

	if !contains(webEnvironment, newEnvString) {
		t.Fatalf("Failed to update existing key FOO. %s does not exist.", newEnvString)
	}

	v2ComposeNoBuildEnv.InjectEnvironmentVariables(environmentVariables)
	svc, _ = v2ComposeNoBuildEnv.project.ServiceConfigs.Get("web")
	webEnvironment = svc.Environment

	if len(webEnvironment) == 0 {
		t.Fatalf("Build section was not created in Compose file without build section. Actual: %v", svc)
	}
}

func TestRewritePortBindings(t *testing.T) {
	var svc *config.ServiceConfig

	setup()

	v1Compose.RewritePortBindings()
	svc, _ = v1Compose.project.ServiceConfigs.Get("web")
	v1Ports := svc.Ports

	if len(v1Ports) != 1 || v1Ports[0] != "8080" {
		t.Fatalf("Failed to rewrite web ports in V1. Expect: [8080], actual: %v", v1Ports)
	}

	v2Compose.RewritePortBindings()
	svc, _ = v2Compose.project.ServiceConfigs.Get("web")
	v2WebPorts := svc.Ports

	if len(v2WebPorts) != 1 || v2WebPorts[0] != "8080" {
		t.Fatalf("Failed to rewrite web ports in V2. Expect: [8080], actual: %v", v2WebPorts)
	}

	svc, _ = v2Compose.project.ServiceConfigs.Get("db")
	v2DbPorts := svc.Ports

	if len(v2DbPorts) != 1 || v2DbPorts[0] != "5432" {
		t.Fatalf("Failed to rewrite non-web ports in V2. Expect: [5432], actual: %v", v2DbPorts)
	}
}

func TestSaveAs(t *testing.T) {
	setup()

	newFilePath := filepath.Join("/tmp", "new-docker-compose.yml")

	if fileExists(newFilePath) {
		os.Remove(newFilePath)
	}

	if err := v1Compose.SaveAs(newFilePath); err != nil {
		t.Fatalf("SaveAs() fails: %s", err.Error())
	}

	if !fileExists(newFilePath) {
		t.Fatalf("SaveAs() does not create %s", newFilePath)
	}

	os.Remove(newFilePath)
}

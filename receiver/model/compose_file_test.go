package model

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	V1FilePath, V2FilePath, V2FilePathBuildArg, V2FilePathNoBuildEnv             string
	V1ComposeFile, V2ComposeFile, V2ComposeFileBuildArg, V2ComposeFileNoBuildEnv *ComposeFile
)

func contains(slice []interface{}, item string) bool {
	set := make(map[interface{}]struct{}, len(slice))

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
	V1FilePath = fixturePath("docker-compose-v1.yml")
	V2FilePath = fixturePath("docker-compose-v2.yml")
	V2FilePathBuildArg = fixturePath("docker-compose-v2-buildarg.yml")
	V2FilePathNoBuildEnv = fixturePath("docker-compose-v2-nobuildenv.yml")
	V1ComposeFile, _ = NewComposeFile(V1FilePath)
	V2ComposeFile, _ = NewComposeFile(V2FilePath)
	V2ComposeFileBuildArg, _ = NewComposeFile(V2FilePathBuildArg)
	V2ComposeFileNoBuildEnv, _ = NewComposeFile(V2FilePathNoBuildEnv)
}

func TestInjectBuildArgs(t *testing.T) {
	var (
		buildArgs      map[string]string
		buildArgString string
		webBuildArgs   []interface{}
	)

	setup()

	buildArgs = map[string]string{
		"FOO": "hoge",
		"BAR": "fuga",
		"BAZ": "piyo",
	}

	// TODO: test for V1

	V2ComposeFile.InjectBuildArgs(buildArgs)
	webService := V2ComposeFile.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})

	webBuildArgs = webService["build"].(map[interface{}]interface{})["args"].([]interface{})

	for key, value := range buildArgs {
		buildArgString = key + "=" + value

		if !contains(webBuildArgs, buildArgString) {
			t.Fatalf("Compose File V2 does not contain %s", key)
		}
	}

	buildArgs = map[string]string{
		"FOO": "hogefugapiyo",
	}
	V2ComposeFileBuildArg.InjectBuildArgs(buildArgs)
	webBuildArgs = V2ComposeFileBuildArg.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["build"].(map[interface{}]interface{})["args"].([]interface{})

	oldBuildArgtring := "FOO=hoge"
	newBuildArgString := "FOO=hogefugapiyo"

	if contains(webBuildArgs, oldBuildArgtring) {
		t.Fatalf("Failed to update existing key FOO. %s still exists.", oldBuildArgtring)
	}

	if !contains(webBuildArgs, newBuildArgString) {
		t.Fatalf("Failed to update existing key FOO. %s does not exist.", newBuildArgString)
	}

	V2ComposeFileNoBuildEnv.InjectBuildArgs(buildArgs)

	if V2ComposeFileNoBuildEnv.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["build"] != nil {
		t.Fatalf("Build section was created in Compose file without build section. Actual: %v", V2ComposeFileNoBuildEnv.Yaml)
	}
}

func TestInjectEnvironmentVariables(t *testing.T) {
	var (
		environmentVariables map[string]string
		envString            string
		webEnvironment       []interface{}
	)

	setup()

	environmentVariables = map[string]string{
		"FOO": "hoge",
		"BAR": "fuga",
		"BAZ": "piyo",
	}

	V1ComposeFile.InjectEnvironmentVariables(environmentVariables)
	webEnvironment = V1ComposeFile.Yaml["web"].(map[interface{}]interface{})["environment"].([]interface{})

	for key, value := range environmentVariables {
		envString = key + "=" + value

		if !contains(webEnvironment, envString) {
			t.Fatalf("Compose File V1 does not contain %s", key)
		}
	}

	if !contains(webEnvironment, "DATABASE_HOST=db") {
		t.Fatalf("Original string DATABASE_HOST=db is dismissed. environments: %v", V1ComposeFile.Yaml["web"].(map[interface{}]interface{})["environment"].([]interface{}))
	}

	V2ComposeFile.InjectEnvironmentVariables(environmentVariables)
	webEnvironment = V2ComposeFile.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["environment"].([]interface{})

	for key, value := range environmentVariables {
		envString = key + "=" + value

		if !contains(webEnvironment, envString) {
			t.Fatalf("Compose File V2 does not contain %s", key)
		}
	}

	environmentVariables = map[string]string{
		"FOO": "hogefugapiyo",
	}

	V2ComposeFile.InjectEnvironmentVariables(environmentVariables)
	webEnvironment = V2ComposeFile.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["environment"].([]interface{})

	oldEnvString := "FOO=hoge"
	newEnvString := "FOO=hogefugapiyo"

	if contains(webEnvironment, oldEnvString) {
		t.Fatalf("Failed to update existing key FOO. %s still exists.", oldEnvString)
	}

	if !contains(webEnvironment, newEnvString) {
		t.Fatalf("Failed to update existing key FOO. %s does not exist.", newEnvString)
	}

	V2ComposeFileNoBuildEnv.InjectEnvironmentVariables(environmentVariables)

	if V2ComposeFileNoBuildEnv.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["environment"] == nil {
		t.Fatalf("Build section was not created in Compose file without build section. Actual: %v", V2ComposeFileNoBuildEnv.Yaml)
	}
}

func TestRewritePortBindings(t *testing.T) {
	setup()

	V1ComposeFile.RewritePortBindings()
	v1Ports := V1ComposeFile.Yaml["web"].(map[interface{}]interface{})["ports"].([]interface{})

	if len(v1Ports) != 1 || v1Ports[0] != "8080" {
		t.Fatalf("Failed to rewrite web ports in V1. Expect: [8080], actual: %v", v1Ports)
	}

	V2ComposeFile.RewritePortBindings()
	v2WebPorts := V2ComposeFile.Yaml["services"].(map[interface{}]interface{})["web"].(map[interface{}]interface{})["ports"].([]interface{})

	if len(v2WebPorts) != 1 || v2WebPorts[0] != "8080" {
		t.Fatalf("Failed to rewrite web ports in V2. Expect: [8080], actual: %v", v2WebPorts)
	}

	v2DbPorts := V2ComposeFile.Yaml["services"].(map[interface{}]interface{})["db"].(map[interface{}]interface{})["ports"].([]interface{})

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

	if err := V1ComposeFile.SaveAs(newFilePath); err != nil {
		t.Fatalf("SaveAs() fails: %s", err.Error())
	}

	if !fileExists(newFilePath) {
		t.Fatalf("SaveAs() does not create %s", newFilePath)
	}

	os.Remove(newFilePath)
}

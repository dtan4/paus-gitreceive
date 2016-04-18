package main

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	V1FilePath, V2FilePath       string
	V1ComposeFile, V2ComposeFile *ComposeFile
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

func setup() {
	workingDir, _ := os.Getwd()

	V1FilePath = filepath.Join(workingDir, "fixtures", "docker-compose-v1.yml")
	V2FilePath = filepath.Join(workingDir, "fixtures", "docker-compose-v2.yml")
	V1ComposeFile, _ = NewComposeFile(V1FilePath)
	V2ComposeFile, _ = NewComposeFile(V2FilePath)
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
	envString = "FOO=hogefugapiyo"

	if !contains(webEnvironment, envString) {
		t.Fatalf("Failed to update existing key FOO. expected: hogefugapiyo, actual: %v", webEnvironment)
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

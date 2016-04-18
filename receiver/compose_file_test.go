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

func setup() {
	workingDir, _ := os.Getwd()

	V1FilePath = filepath.Join(workingDir, "fixtures", "docker-compose-v1.yml")
	V2FilePath = filepath.Join(workingDir, "fixtures", "docker-compose-v2.yml")
	V1ComposeFile, _ = NewComposeFile(V1FilePath)
	V2ComposeFile, _ = NewComposeFile(V2FilePath)
}

func TestIsVersion2(t *testing.T) {
	setup()

	if V1ComposeFile.IsVersion2() {
		t.Fatalf(V1FilePath + " is actually not based on Compose File Version 2.")
	}

	if !V2ComposeFile.IsVersion2() {
		t.Fatalf(V2FilePath + " is actually based on Compose File Version 2.")
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsVersion2(t *testing.T) {
	workingDir, _ := os.Getwd()

	v1FilePath := filepath.Join(workingDir, "fixtures", "docker-compose-v1.yml")
	v1ComposeFile, _ := NewComposeFile(v1FilePath)

	if v1ComposeFile.IsVersion2() {
		t.Fatalf(v1FilePath + " is actually not based on Compose File Version 2.")
	}

	v2FilePath := filepath.Join(workingDir, "fixtures", "docker-compose-v2.yml")
	v2ComposeFile, _ := NewComposeFile(v2FilePath)

	if !v2ComposeFile.IsVersion2() {
		t.Fatalf(v2FilePath + " is actually based on Compose File Version 2.")
	}
}

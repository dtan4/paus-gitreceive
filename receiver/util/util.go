package util

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

	"github.com/pkg/errors"
)

func printLine(r io.Reader) {
	sc := bufio.NewScanner(r)

	for sc.Scan() {
		fmt.Println("       " + sc.Text())
	}
}

func GetSubmodules(repositoryPath string) error {
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

func RemoveUnpackedFiles(repositoryPath, newComposeFilePath string) error {
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

func RunCommand(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating stdout failed. command: %v", cmd.Args))
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating stderr failed. command: %v", cmd.Args))
	}

	cmd.Start()

	go printLine(stdout)
	go printLine(stderr)

	if err = cmd.Wait(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("command execution failed. command: %v", cmd.Args))
	}

	return nil
}

func UnpackReceivedFiles(repositoryDir, username, projectName string, stdin io.Reader) (string, error) {
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

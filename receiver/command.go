package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

func printLine(r io.Reader) {
	sc := bufio.NewScanner(r)

	for sc.Scan() {
		fmt.Println("       " + sc.Text())
	}
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

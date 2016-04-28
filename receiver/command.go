package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func printLine(r io.Reader) {
	sc := bufio.NewScanner(r)

	for sc.Scan() {
		fmt.Println(sc.Text())
	}
}

func RunCommand(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	cmd.Start()

	go printLine(stdout)
	go printLine(stderr)

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

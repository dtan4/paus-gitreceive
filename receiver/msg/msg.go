package msg

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Println prints the given message in stderr
func Println(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// Printf prints the given message with specified format in stderr
func Printf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// PrintTitle prints title with bold string
func PrintTitle(title string) {
	c := color.New(color.Bold)
	c.EnableColor()
	fmt.Fprintln(os.Stderr, c.SprintfFunc()("======> %s", title))
}

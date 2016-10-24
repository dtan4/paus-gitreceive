package msg

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	bold = color.New(color.Bold)
	red  = color.New(color.FgRed)
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
	bold.EnableColor()
	fmt.Fprintln(os.Stderr, bold.SprintfFunc()("======> %s", title))
}

// PrintError prints error with red string
func PrintError(s string) {
	red.EnableColor()
	fmt.Fprintln(os.Stderr, red.SprintfFunc()("%s", s))
}

// PrintError prints error with specified format and red string
func PrintErrorf(format string, a ...interface{}) {
	red.EnableColor()
	fmt.Fprintln(os.Stderr, red.SprintfFunc()(format, a...))
}

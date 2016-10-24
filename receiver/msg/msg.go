package msg

import (
	"fmt"
	"os"
)

// Println prints the given message in stderr
func Println(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// Printf prints the given message with specified format in stderr
func Printf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

package main

import (
	"fmt"
)

var (
	GoVersion string
	Revision  string
	Version   string
)

func printVersion() {
	fmt.Println("Version:   " + Version)
	fmt.Println("Revision:  " + Revision)
	fmt.Println("GoVersion: " + GoVersion)
}

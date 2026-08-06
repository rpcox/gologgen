package main

import (
	"fmt"
	"os"
)

var (
	usageText = `

	NAME

	SYNOPSIS

	DESCRIPTION

	EXAMPLES
	`
)

func Usage(b bool, exitCode int) {
	if b {
		fmt.Println(usageText)
		os.Exit(exitCode)
	}
}

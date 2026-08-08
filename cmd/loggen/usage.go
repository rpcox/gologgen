package main

import (
	"fmt"
)

var (
	usageText = `

	NAME

	SYNOPSIS

	DESCRIPTION

	EXAMPLES
	`
)

func Usage() {
	fmt.Println(usageText)
}

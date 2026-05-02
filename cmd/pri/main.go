package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	lg "github.com/rpcox/gologgen/pkg/loggen"
	"github.com/rpcox/pkg/exit"
)

var (
	IsAllDigits = regexp.MustCompile(`^\d+$`)
	IsDottedStr = regexp.MustCompile(`^\w+\.\w+$`)
	IsHelp      = regexp.MustCompile(`^\-?h(elp)?`)
)

func main() {
	exit.If(len(os.Args) < 2, fmt.Errorf("try 'pri -help'"), 1)

	if IsHelp.MatchString(os.Args[1]) {
		Usage(0)
	} else if os.Args[1] == `list` {
		fmt.Printf(" Code   Facility\n\n")
		for value, name := range lg.GetFacilityNames() {
			fmt.Printf("%3d     %s\n", value, name)
		}
		fmt.Printf("\n Code   Severity\n\n")
		for value, name := range lg.GetSeverityNames() {
			fmt.Printf("%3d     %s\n", value, name)
		}
		fmt.Println()
	} else if IsAllDigits.MatchString(os.Args[1]) {
		i, err := strconv.Atoi(os.Args[1])
		exit.If(err != nil, err, 1)
		pri, err := lg.PriValItoa(i)
		exit.If(err != nil, err, 1)
		fmt.Println(pri)
	} else if IsDottedStr.MatchString(os.Args[1]) {
		pri, err := lg.PriValAtoi(os.Args[1])
		exit.If(err != nil, err, 1)
		fmt.Println(pri)
	} else {
		Usage(1)
	}
}

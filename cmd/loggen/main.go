// A tool to generate syslog records
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	lg "github.com/rpcox/gologgen/pkg/loggen"
	//"github.com/rpcox/pkg/exit"
)

const _version = "0.0.1"
const _tool = "loggen"

// The information necessary to send data down range
type Loggen struct {
	Server     string  // destination server
	Port       int     // destionation port
	Proto      string  // protocol udp || tcp
	PRI        int     // syslog priority
	Message    *string // random string message
	Count      int     // count of records to send
	GoRoutines int     // the number of go routines to initiate
}

// Build option to track git commit/build if desired
func Version(b bool) {
	if b {
		fmt.Fprintf(os.Stdout, "%s v%s\n", _tool, _version)
		fmt.Fprintf(os.Stdout, " - %s\n", lg.GetVersion())
		os.Exit(lg.Success)
	}
}

var (
	opts = lg.Options{}
)

func main() {
	// Destination
	flag.StringVar(&opts.Destination.Dst, "dst", "", "Specify the destination [by name, IP or file]")
	flag.StringVar(&opts.Destination.Port, "dport", "514", "Specify the destination port")
	// Operation
	flag.BoolVar(&opts.Operation.Bsd, "bsd", false, "Use RFC 3164 (BSD) format")
	flag.IntVar(&opts.Operation.Count, "count", 1, "The number of messages to send to the destination")
	flag.BoolVar(&opts.Operation.Udp, "udp", false, "Use UDP for transport")
	flag.BoolVar(&opts.Operation.Tls, "tls", false, "Use TLS for transport")
	flag.IntVar(&opts.Operation.MsgLen, "msglen", 128, "Specify the length of the random message")
	flag.IntVar(&opts.Operation.GoRoutines, "gr", 1, "Specify the number of Go routines")
	flag.StringVar(&opts.Pri, "pri", "local0.info", "Specify the priority [facility.severity]")
	flag.BoolVar(&opts.Stats, "stats", false, "Display EPS stats")
	// Record
	flag.IntVar(&opts.Record.Version, "v", 1, "Specify the RFC 5424 record version")
	flag.StringVar(&opts.Record.AppName, "appname", "loggen", "Specify APPNAME (application name) field")
	flag.StringVar(&opts.Record.MsgId, "msgid", "-", "Specify MSGID (message id) field")
	flag.StringVar(&opts.Record.Sd, "sd", "-", "Specify SD (structured data) field")
	// Util
	flag.BoolVar(&opts.Debug, "debug", false, "Enable verbose logging to console on stderr")
	_help := flag.Bool("help", false, "Display usage and exit")
	_version := flag.Bool("version", false, "Diplay version and exit")
	flag.Parse()

	Version(*_version)
	Usage(*_help, lg.Success)

	start := time.Now()
	lgen := lg.NewLoggen(&opts)
	err := lgen.Exec()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	lgen.Close()
	fmt.Printf("main: elapsed: %v\n", time.Since(start))
}

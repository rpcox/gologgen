// A tool to generate RFC 3164 and RFC 5424 formatted syslog records
package main

import (
	"flag"
	"fmt"
	"os"

	lg "github.com/rpcox/gologgen/pkg/loggen"
	"github.com/rpcox/pkg/exit"
)

const _version = "0.2.0"
const _tool = "loggen"

func Version(b bool) {
	if b {
		fmt.Fprintf(os.Stdout, "%s v%s\n", _tool, _version)
		fmt.Fprintf(os.Stdout, " - %s\n", lg.GetVersion())
		os.Exit(lg.Success)
	}
}

var (
	bsd     = flag.NewFlagSet("bsd", flag.ExitOnError)
	ietf    = flag.NewFlagSet("ietf", flag.ExitOnError)
	help    = flag.NewFlagSet("help", flag.ExitOnError)
	version = flag.NewFlagSet("version", flag.ExitOnError)
	opts    = lg.Options{}

	subCmds = map[string]*flag.FlagSet{
		bsd.Name():     bsd,
		ietf.Name():    ietf,
		help.Name():    help,
		version.Name(): version,
	}
)

func SetCommonFlags() {
	for _, fs := range subCmds {
		fs.StringVar(&opts.Destination.Dst, "dst", "", "Specify the destination [by name or IP]")
		fs.IntVar(&opts.Destination.Port, "dport", 514, "Specify the destination port")
		fs.BoolVar(&opts.Operation.Udp, "udp", false, "Use UDP for transport")
		fs.BoolVar(&opts.Operation.Tls, "tls", false, "Use TLS for transport")
		fs.IntVar(&opts.Operation.Count, "count", 1, "The number of messages to send to the destination")
		fs.IntVar(&opts.Operation.Workers, "w", 1, "Specify the number of client workers")
		fs.StringVar(&opts.Operation.Duration, "duration", "", "Specify the run duration [e.g. 1h30m5s]")
		fs.StringVar(&opts.Pri, "pri", "local0.info", "Specify the priority [facility.severity]")
		fs.BoolVar(&opts.Stats, "stats", false, "Display EPS stats")
		fs.BoolVar(&opts.Debug, "debug", false, "Enable verbose logging to console on stderr")
		fs.IntVar(&opts.Operation.MsgLen, "msglen", 128, "Specify the length of the random message")
	}
}

func SetSpecificFlags() {
	for _, fs := range subCmds {
		if fs.Name() != `help` || fs.Name() != `version` {
			if fs.Name() == `ietf` {
				fs.IntVar(&opts.Record.Version, "v", 1, "Specify the RFC 5424 record version")
				fs.StringVar(&opts.Record.AppName, "appname", "loggen", "Specify APPNAME (application name) field")
				fs.StringVar(&opts.Record.MsgId, "msgid", "-", "Specify MSGID (message id) field")
				fs.StringVar(&opts.Record.Sd, "sd", "-", "Specify SD (structured data) field")

			} else if fs.Name() == `bsd` {
				fs.StringVar(&opts.Record.AppName, "tag", "tag", "Specify APPNAME (application name) field")
			}
		}
	}

}

func CheckCmdLine(help, version bool) string {
	if len(os.Args) == 1 { // nothing on cmd line
		fmt.Printf("expecting sub command. try '%s help'", _tool)
		os.Exit(lg.ErrInsufficientArg)
	}

	cmd, ok := subCmds[os.Args[1]]
	if !ok {
		fmt.Println(ok)
		Version(version)
		if help {
			Usage()
			os.Exit(lg.Success)
		}
	}

	return cmd.Name()
}

func main() {
	SetCommonFlags()
	SetSpecificFlags()
	flag.Usage = Usage
	_help := flag.Bool("help", false, "Display usage and exit")
	_version := flag.Bool("version", false, "Diplay version and exit")
	flag.Parse()

	subCommand := CheckCmdLine(*_help, *_version)
	switch subCommand {
	case `bsd`:
		opts.Operation.Bsd = true
		bsd.Parse(os.Args[2:])
	case `ietf`:
		ietf.Parse(os.Args[2:])
	case `help`:
		Usage()
		os.Exit(lg.Success)
	case `version`:
		Version(true)
	}

	exit.If(opts.Operation.Tls, fmt.Errorf("-tls not implemented"), lg.ErrNotImplemented)
	lgen := lg.NewLoggen(&opts)
	err := lgen.Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	lgen.Close()
}

// SDG

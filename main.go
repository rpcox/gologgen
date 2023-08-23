// A tool to generate syslog records
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rpcox/gologgen/syslog"
)

const _version = "0.4"
const _tool = "gologgen"

var (
	_commit string
	_branch string
)

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
		if _commit != "" {
			// go build -ldflags="-X main._commit=$(git rev-parse --short HEAD) -X main._branch=$(git branch | awk '{print $2}')"
			fmt.Fprintf(os.Stdout, "%s v%s (commit: %s, branch: %s)\n", _tool, _version, _commit, _branch)
		} else {
			// go build
			fmt.Fprintf(os.Stdout, "%s v%s\n", _tool, _version)
		}
		os.Exit(0)
	}
}

// Short circuit logic for quick exit
func ExitUnless(b bool, s string) {
	if !b {
		log.Fatal(s)
	}
}

// Usage statement
func Usage(b bool) {
	if b {
		doc := `
  NAME
  	gologgen - syslog record generator

  SYNOPSIS
  	gologgen [OPTIONS]

  DESCRIPTION
	Generate syslog RFC3164 or RFC5424 traffic

  OPTIONS
 
 `
		fmt.Println(doc)
		flag.PrintDefaults()
		os.Exit(0)
	}
}

// Generate a random string of A-Z chars with len = l
func RandomString(len int) *string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	s := string(bytes)
	return &s
}

// Validate the protocol request
func ValidateProtocol(udp bool, tcp bool, tls bool) (string, error) {
	s := "-tcp, -tls, or -udp.  Only one protocol may be selected"
	if udp && tcp {
		return "", fmt.Errorf("%s", s)
	} else if udp && tls {
		return "", fmt.Errorf("%s", s)
	} else if tcp && tls {
		return "", fmt.Errorf("%s", s)
	}

	if tcp {
		return "tcp", nil
	} else if udp {
		return "udp", nil
	} else {
		return "tls", fmt.Errorf("%s", "not implemented")
	}
}

// Go routine launch point to send IETF records
func SendIetfRecords(lg *Loggen, ietf *syslog.RFC5424) {
	var wg sync.WaitGroup
	dst := fmt.Sprintf("%s:%d", lg.Server, lg.Port)
	for id := 1; id <= lg.GoRoutines; id++ {
		wg.Add(1)
		go SendIetf(dst, *ietf.Format, lg.Proto, *lg.Message, id, lg.Count, &wg)
	}

	wg.Wait()
}

// Send the IETF records to the destination
func SendIetf(dst, format, proto, msg string, id, count int, wg *sync.WaitGroup) {
	conn, err := net.Dial(proto, dst)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	t0 := time.Now()
	for i := 1; i <= count; i++ {
		date := time.Now().UTC().Format("2006-01-02T03:04:05.00Z")
		s := fmt.Sprintf(format, date, msg)
		conn.Write([]byte(s))
	}
	t1 := time.Now()
	fmt.Println("Go routine [", id, "] completed. elapsed", t1.Sub(t0))
	wg.Done()
}

// Go routine launch point to send BSD records
func SendBsdRecords(lg *Loggen, bsd *syslog.RFC3164) {
	var wg sync.WaitGroup
	dst := fmt.Sprintf("%s:%d", lg.Server, lg.Port)
	for id := 1; id <= lg.GoRoutines; id++ {
		wg.Add(1)
		if bsd.RFC3339 {
			go SendRFC3339Bsd(dst, *bsd.Format, lg.Proto, *lg.Message, id, lg.Count, &wg)
		} else {
			go SendBsd(dst, *bsd.Format, lg.Proto, *lg.Message, id, lg.Count, &wg)
		}
	}

	wg.Wait()
}

// Send the BSD records to the destination
func SendBsd(dst, format, proto, msg string, id, count int, wg *sync.WaitGroup) {
	conn, err := net.Dial(proto, dst)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	t0 := time.Now()
	for i := 1; i <= count; i++ {
		date := time.Now().UTC().Format(time.Stamp)
		s := fmt.Sprintf(format, date, msg)
		conn.Write([]byte(s))
	}
	t1 := time.Now()
	fmt.Println("Go routine [", id, "] completed. elapsed", t1.Sub(t0))
	wg.Done()
}

// Send the BSD records w/ RFC3339 timestamp to the destination
func SendRFC3339Bsd(dst, format, proto, msg string, id, count int, wg *sync.WaitGroup) {
	conn, err := net.Dial(proto, dst)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	t0 := time.Now()
	for i := 1; i <= count; i++ {
		date := time.Now().UTC().Format(time.RFC3339)
		s := fmt.Sprintf(format, date, msg)
		conn.Write([]byte(s))
	}
	t1 := time.Now()
	fmt.Printf("go routine[%d] completed. %d records. time elapsed: %v\n", id, count, t1.Sub(t0))
	wg.Done()
}

func main() {
	bsd := new(syslog.RFC3164)
	ietf := new(syslog.RFC5424)
	lg := new(Loggen)

	flag.StringVar(&lg.Server, "server", "", "Specify the destination server (by name or IP)")
	flag.IntVar(&lg.Port, "port", 514, "Specify the destination port")
	flag.IntVar(&lg.Count, "count", 1, "The number of messages to send to the destination server")
	// RFC3164
	rfc3164 := flag.Bool("rfc3164", false, "Specify RFC3164 format. Default format is RFC5424")
	flag.BoolVar(&bsd.PID, "pid", false, "RFC3164: Insert PID with tag ( e.g., TAG[PID] )")
	flag.BoolVar(&bsd.RFC3339, "rfc3339", false, "RFC3164: Use RFC3339 time format.")
	flag.StringVar(&bsd.Tag, "tag", "gologgen", "RFC3164: Specify the RFC3164 tag in the syslog record")
	// RFC5424
	flag.StringVar(&ietf.AppName, "appname", "gologgen", "RFC5424: Use the specified tag (3164) or AppName (5424) in the syslog record")
	flag.BoolVar(&ietf.ProcID, "procid", false, "RFC5424: Specify the PROCID")
	flag.StringVar(&ietf.MsgID, "msgid", "-", "RFC5424: Specify the MSGID")
	flag.StringVar(&ietf.Sd, "sd", "-", "RFC5424: Specify the structured data")

	contentLength := flag.Int("msg-length", 64, "Set the random message to this length")
	goroutines := flag.Int("gr", 1, "Specify the number of Go routines to initiate")
	priority := flag.String("priority", "local0.info", "Set the specified priority for the syslog record")

	udp := flag.Bool("udp", false, "Use UDP")
	tcp := flag.Bool("tcp", false, "Use TCP")
	tls := flag.Bool("tls", false, "Use TLS (not implemented)")

	help := flag.Bool("help", false, "Display usage and exit")
	version := flag.Bool("version", false, "Diplay version and exit")

	flag.Parse()
	Version(*version)
	Usage(*help || flag.NFlag() < 1)
	ExitUnless(len(lg.Server) > 0, "-server is required")
	ExitUnless((lg.Port > 0) && (lg.Port < 65535), "-port must > 0 and <= 65535")
	ExitUnless(lg.Count > 0, "-count must be > 0")
	ExitUnless(*goroutines > 0, "-gr must > 0")
	ExitUnless(*contentLength > 0, "-msg-length must > 0")

	pri, err := syslog.SetPriority(*priority)
	if err != nil {
		log.Fatal(err)
	}
	if lg.Proto, err = ValidateProtocol(*udp, *tcp, *tls); err != nil {
		log.Fatal(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "FIXTHIS"
	}

	lg.GoRoutines = *goroutines
	lg.Message = RandomString(*contentLength)
	fmt.Printf("Sending %d syslog records down range\n", *goroutines*lg.Count)

	if *rfc3164 {
		bsd.PRI = pri
		bsd.Hostname = hostname
		syslog.SetBSDRecordFormat(bsd)
		SendBsdRecords(lg, bsd)
	} else {
		ietf.PRI = pri
		ietf.Version = 1
		ietf.Hostname = hostname
		syslog.SetIETFRecordFormat(ietf)
		SendIetfRecords(lg, ietf)
	}

	fmt.Println("Done")
}

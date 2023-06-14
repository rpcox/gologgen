// A tool to generate syslog data
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const _version = "0.2"
const _tool = "gologgen"

var (
	_commit string
	_branch string

	Facility = map[string]int{
		"kernel":   0,
		"user":     1,
		"mail":     2,
		"daemon":   3,
		"auth":     4,
		"syslog":   5,
		"lpr":      6,
		"news":     7,
		"uucp":     8,
		"cron":     9,
		"authpriv": 10,
		"ftp":      11,
		"ntp":      12,
		"log1":     13,
		"log2":     14,
		"clock":    15,
		"local0":   16,
		"local1":   17,
		"local2":   18,
		"local3":   19,
		"local4":   20,
		"local5":   21,
		"local6":   22,
		"local7":   23,
	}

	Severity = map[string]int{
		"emerg":    0,
		"alert":    1,
		"critical": 2,
		"error":    3,
		"warn":     4,
		"notice":   5,
		"info":     6,
		"debug":    7,
	}
)

type Loggen struct {
	Dest string
	//File string
	Host  string
	PID   int
	Port  int
	PRI   int
	Tag   string
	Pad   *string
	Count int
	Proto string
}

func Version(v bool) {
	if v {
		fmt.Fprintf(os.Stdout, "%s v%s\n", _tool, _version)
		os.Exit(0)
	}
}

func Usage(b bool) {
	if b {
		doc := `
  NAME
  	gologgen - syslog record generator

  SYNOPSIS
  	gologgen [OPTIONS]

  DESCRIPTION
	Generate syslog traffic

  OPTIONS
 
 `
		fmt.Println(doc)
		flag.PrintDefaults()
		os.Exit(0)
	}
}

// RandomString - Generate a random string of A-Z chars with len = l
func RandomString(len int) *string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	s := string(bytes)
	return &s
}

// Set the syslog priority
func setPriority(priority *string) int {
	// priority = facility * 8 + severity
	s := strings.Split(*priority, ".")
	facility, ok_facility := Facility[s[0]]
	severity, ok_severity := Severity[s[1]]

	if ok_facility && ok_severity {
		return facility*8 + severity
	} else {
		log.Fatal("priority '", *priority, "' not supported")
	}

	return 0
}

// Format the record for sending downstream
func FormatRecord(l *Loggen, m *sync.Mutex) string {
	date := time.Now().UTC().Format(time.RFC3339)
	m.Lock()
	s := fmt.Sprintf("<%d>%s %s %s[%d]: %s\n", l.PRI, date, l.Host, l.Tag, l.PID, *l.Pad)
	m.Unlock()
	return s
}

// Send it all down range
func SendIt(l *Loggen, i int, wg *sync.WaitGroup, m *sync.Mutex) {
	conn, err := net.Dial(l.Proto, l.Dest)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	t0 := time.Now()
	for i := 1; i <= l.Count; i++ {
		m := FormatRecord(l, m)
		conn.Write([]byte(m))
	}
	t1 := time.Now()
	fmt.Println("Go routine[", i, "]completed. elapsed", t1.Sub(t0))
	wg.Done()
}

func main() {
	var (
		priority = flag.String("pri", "local0.info", "Set the specified priority for the syslog record")
		//facDump    = flag.Bool("facility-dump", false, "Dump the facility names and values")
		//rfc3339    = flag.Bool("rfc3339", false, "Use RFC 3339 date format")
		server     = flag.String("server", "", "Send the message to the specified server")
		port       = flag.Int("port", 514, "Send the message to the specified destination port")
		mlen       = flag.Int("len", 128, "Set the random message to this length")
		count      = flag.Int("count", 1, "The number of messages to send down range")
		tag        = flag.String("tag", "gologgen", "Use the specified process tag in the syslog record")
		goroutines = flag.Int("gr", 1, "Specify the number of Go routines to initiate")
		tcp        = flag.Bool("tcp", false, "Use TCPv4.  Default is UDPv4")

		help    = flag.Bool("help", false, "Display usage and exit")
		version = flag.Bool("version", false, "Diplay version and exit")
	)

	flag.Parse()
	Version(*version)
	Usage(*help || flag.NFlag() < 1)

	l := Loggen{Dest: *server + ":" + strconv.Itoa(*port)}
	l.Host, _ = os.Hostname()
	l.PID = os.Getpid()
	l.Port = *port
	l.PRI = setPriority(priority)
	l.Tag = *tag
	l.Pad = RandomString(*mlen)
	l.Count = *count

	if *tcp {
		l.Proto = "tcp4"
	} else {
		l.Proto = "udp4"
	}

	var wg sync.WaitGroup
	var m sync.Mutex
	for i := 1; i <= *goroutines; i++ {
		wg.Add(1)
		go SendIt(&l, i, &wg, &m)
	}

	wg.Wait()
	fmt.Println("All go routines have completed execution")
}

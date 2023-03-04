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
	"time"
)

var (
	//debug    = flag.Bool("debug", false, "Turn on debugging")
	priority = flag.String("pri", "local0.info", "Send the message with the specified priority")
	server   = flag.String("s", "TAG", "Use the specified process tag")
	port     = flag.Int("port", 514, "Send the message to the specified destination port")
	pad      = flag.Int("pad", 128, "Set the random message to this length")
	count    = flag.Int("count", 1, "Send this many messages down range")
	tag      = flag.String("tag", "TAG", "Use the specified process tag")

	help    = flag.Bool("help", false, "Display usage and exit")
	version = flag.Bool("version", false, "Diplay version and exit")

	Facility = map[string]int{
		"local0": 16,
		"local1": 17,
		"local2": 18,
		"local3": 19,
		"local4": 20,
		"local5": 21,
		"local6": 22,
		"local7": 23,
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
	Dst string
	//File string
	Host string
	PID  int
	Port int
	PRI  int
	Tag  string
	Pad  *string
}

func ShowVersion() {
	ver := "0.1"
	fmt.Println("gologgen v", ver)
	os.Exit(0)
}

func Usage() {
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

// RandomString - Generate a random string of A-Z chars with len = l
func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

func Initialize(l *Loggen) error {

	// priority = facility * 8 + severity
	s := strings.Split(*priority, ".")
	facility, okf := Facility[s[0]]
	severity, oks := Severity[s[1]]

	if okf && oks {
		l.PRI = facility*8 + severity
	} else {
		err := fmt.Errorf("'%s' not supported", *priority)
		return err
	}

	//l.File = *file
	l.Host, _ = os.Hostname()
	l.PID = os.Getpid()
	l.Port = *port
	l.Dst = *server + ":" + strconv.Itoa(*port)
	l.Tag = *tag

	p := RandomString(*pad)
	l.Pad = &p

	return nil
}

func FormatRecord(l *Loggen) string {
	date := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf("<%d>%s %s %s[%d]: %s\n", l.PRI, date, l.Host, l.Tag, l.PID, *l.Pad)
}

func NetWrite(nc net.Conn, l *Loggen) int {
	m := FormatRecord(l)
	nc.Write([]byte(m))
	return 1
}

func main() {

	flag.Parse()
	if *version {
		ShowVersion()
	} else if *help || flag.NFlag() < 1 {
		Usage()
	}

	var loggen Loggen
	err := Initialize(&loggen)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("tcp", loggen.Dst)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for i := 1; i <= *count; i++ {
		NetWrite(conn, &loggen)
	}
}

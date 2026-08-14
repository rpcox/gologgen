package loggen

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rpcox/pkg/exit"
)

const (
	tool    = "pkg-loggen"
	version = `v0.2.0`
)

var (

	// RFC 5424 6.2.1 p10
	Facility = map[string]int{
		"kernel":   0,  // kernel messages
		"user":     1,  // user level messages
		"mail":     2,  // mail system
		"daemon":   3,  // system daemons
		"auth":     4,  // security/authorization messages
		"syslog":   5,  // messages generated internally by syslogd
		"lpr":      6,  // line printer subsystem
		"news":     7,  // network news subsystem
		"uucp":     8,  // uucp subsystem
		"cron":     9,  // clock daemon
		"authpriv": 10, // security/authorization messages
		"ftp":      11, // FTP daemon
		"ntp":      12, // NTP subsystem
		"audit":    13, // log audit
		"alert":    14, // log alert
		"clock":    15, // clock daemon
		"local0":   16, // local use 0 (local0)
		"local1":   17, // local use 1 (local1)
		"local2":   18, // local use 2 (local2)
		"local3":   19, // local use 3 (local3)
		"local4":   20, // local use 4 (local4)
		"local5":   21, // local use 5 (local5)
		"local6":   22, // local use 6 (local6)
		"local7":   23, // local use 7 (local7)
	}

	// RFC 5424 6.2.1 p11
	Severity = map[string]int{
		"emerg":  0, // Emergency: system is unusable
		"alert":  1, // Alert: action must be taken immediately
		"crit":   2, // Critical: critical conditions
		"error":  3, // Error: error conditions
		"warn":   4, // Warning: warning conditions
		"notice": 5, // Notice: normal but significant condition
		"info":   6, // Informational: informational messages
		"debug":  7, // Debug: debug-level messages
	}

	_debug     = false
	timeFormat = `2006-01-02 15:04:05.000000`
)

func GetVersion() string {
	return fmt.Sprintf("%s %s", tool, version)
}

func GetFacilityNames() []string {
	return mapToSlice(Facility)
}

func GetSeverityNames() []string {
	return mapToSlice(Severity)
}

// Receiving priority as a dot separated strings (e.g., 'local0.info'), convert to an integer
// On success return PRI, nil
// On fail, return -1, err
func PriValAtoi(priority string) (int, error) {
	sl := strings.Split(priority, ".")
	if len(sl) < 2 {
		return -1, fmt.Errorf("priority string format. expected facility.severity")
	}

	f := strings.ToLower(sl[0])
	s := strings.ToLower(sl[1])
	facility, facility_exists := Facility[f]
	severity, severity_exists := Severity[s]

	if facility_exists {
		if severity_exists {
			// RFC5424 6.2.1 p11
			return (facility << 3) + severity, nil
		}
		return -1, fmt.Errorf("severity '%s' not supported", sl[1])
	}

	return -1, fmt.Errorf("facility '%s' not supported", sl[0])
}

// Used to invert Facility and Severity maps
func invertMap(m map[string]int) map[int]string {
	im := make(map[int]string, len(m))
	for k, v := range m {
		im[v] = k
	}
	return im
}

// Receiving priority as an integer, convert string w/ format facility.severity
// On success return facility.severity, nil
// On fail, return “, err
func PriValItoa(priority int) (string, error) {
	if priority < 0 || priority > 191 {
		return ``, fmt.Errorf("priority ('%d') > 191. see rfc5424 section 6", priority)
	}

	iFacility := invertMap(Facility)
	iSeverity := invertMap(Severity)

	facility := priority >> 3
	severity := priority & 7

	fs, facility_exists := iFacility[facility]
	ss, severity_exists := iSeverity[severity]

	if facility_exists {
		if severity_exists {
			return strings.Join([]string{fs, ss}, `.`), nil
		} else {
			return ``, fmt.Errorf("severity '%s' does not exist. see rfc5424 6.2.1", fs)
		}
	} else {
		return ``, fmt.Errorf("facility '%s' does not exist. see rfc5424 6.2.1", fs)
	}
}

type Record struct {
	Pri     int
	Version int
	//TimeStamp string  -- just a reminder that this field is in the record. applied elsewhere using type Ts
	Hostname string
	AppName  string
	Pid      string
	MsgId    string
	Sd       string
	Msg      *string
}

type Destination struct {
	Dst      string // from -dst
	DstType  string // stdout, name, IP
	Port     int
	Protocol string
}

type Operation struct {
	Bsd          bool   // -bsd
	Count        int    // -count
	Workers      int    // -w
	MsgLen       int    // -msglen
	Tls          bool   // -tls
	Udp          bool   // -udp
	Duration     string // -duration
	timeDuration time.Duration
}

type Options struct {
	Debug       bool
	Record      Record
	Destination Destination
	Operation   Operation
	Pri         string
	Stats       bool
}

type Loggen struct {
	format     string
	TimeFormat string
	RecordTmpl string
	opt        *Options
	statsChan  chan Stats
	statsSlice []Stats
	start      time.Time
}

type Ts struct {
	TimeStamp string
}

type Interrupt struct {
	mu  sync.Mutex
	yes bool
	at  string
}

func NewLoggen(opts *Options) *Loggen {
	var lg Loggen
	lg.opt = opts
	_debug = lg.opt.Debug
	err := validatePort(opts.Destination.Port) // must be valid port (0 - 65535)
	exit.If(err != nil, err, ErrPortRange)

	//	T, err := validateWDst(opts.Destination.Dst) // file, stdout, name or IP
	//	exit.If(err != nil, err, ErrHostname)
	//	lg.opt.Destination.DstType = T

	lg.format = setFormat(lg.opt.Operation.Bsd) // bsd or ietf

	p, err := setProtocol(lg.opt.Operation) // udp, tcp or tls
	exit.If(err != nil, err, ErrProtocol)
	lg.opt.Destination.Protocol = p

	err = validateCount(lg.opt.Operation.Count) // can't be < 0
	exit.If(err != nil, err, ErrCount)

	t, err := validateDuration(lg.opt.Operation.Duration) // can't be < 0
	exit.If(err != nil, err, ErrCount)
	lg.opt.Operation.timeDuration = t

	lg.opt.Record.Msg = setMessage(lg.opt.Operation.MsgLen) // gen rand string for now

	j, err := setPriority(lg.opt.Pri) // convert 'facility.severity' to int
	exit.If(err != nil, err, ErrPriority)
	lg.opt.Record.Pri = j

	checkConcurrency(lg.opt.Operation.Workers) // give heads up if -gr > GOMAXPROCS

	lg.opt.Record.Pid = strconv.Itoa(os.Getpid())
	h, err := os.Hostname()
	if err != nil {
		h = tool + `host`
		fmt.Fprintf(os.Stderr, "%%note: undetermined host name. setting hostname = '%s'\n", h)
	}

	lg.opt.Record.Hostname = h
	lg.RecordTmpl = setTemplate(lg.opt.Operation.Bsd, lg.format, lg.opt.Record)
	lg.TimeFormat = setTimeFormat(lg.opt.Operation.Bsd)
	lg.statsChan = make(chan Stats, lg.opt.Operation.Workers+1)

	return &lg
}

// Execute the request
func (lg *Loggen) Exec() error {
	var wg sync.WaitGroup
	var client ClientFunc

	switch lg.opt.Destination.Protocol {
	case `udp`:
		client = udpClient
	case `tcp`:
		client = tcpClient
	case `tls`:
		client = tlsClient
	}

	var intr Interrupt
	lg.start = time.Now()
	fmt.Printf(" Starting: %s\n", lg.start.UTC().Format(timeFormat))
	for i := range lg.opt.Operation.Workers {
		wg.Add(1)
		go client(i+1, *lg, &intr, &wg)
	}

	wg.Wait()
	close(lg.statsChan)
	if intr.yes {
		fmt.Printf("\nInterrupt: %s\n", intr.at)
	}
	fmt.Printf("  Elapsed: %v\n", time.Since(lg.start))

	for stat := range lg.statsChan {
		lg.statsSlice = append(lg.statsSlice, stat)
	}

	return nil
}

func (lg *Loggen) Close() {
	lg.presentStats()
}

// SDG

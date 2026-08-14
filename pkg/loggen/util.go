package loggen

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

const (
	Success = iota
	ErrInsufficientArg
	ErrNotImplemented
	ErrHostname
	ErrPortRange
	ErrProtocol
	ErrCount
	ErrPriority
	ErrTemplate
	ErrResolveAddrFail
	ErrUdpConn
	ErrUdpClientFail
	ErrTcpClientFail
	ErrTlsClientFail
	ErrNetworkWrite
)

// If debug enabled, carp a message to stderr
func Carp(s string) {
	if _debug {
		fmt.Fprintf(os.Stderr, "%s\n", s)
	}
}

// Convert map[abc]N to slice[N] = abc
func mapToSlice(m map[string]int) []string {
	sl := make([]string, len(m))
	for name, value := range m {
		sl[value] = name
	}

	return sl
}

// Validate the duration string
func validateDuration(d string) (time.Duration, error) {
	if d == "" {
		return 0, nil
	}

	t, err := time.ParseDuration(d)
	if err != nil {
		var null time.Duration
		return null, err
	}

	return t, nil
}

// Validate the supplied port is within range
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%%err: port value '%d' out of range", port)
	}

	return nil
}

// RFC 3164 (BSD) or RFC 5424 (IETF) format. The string formatter is used to
// generate a text template that can be used to increment date/time
func setFormat(bsd bool) string {
	if bsd {
		// Using RFC 3164 (BSD) format
		//      TIMESTAMP
		// <PRI>DATE TIME HOSTNAME APPNAME[PID]: MESSAGE
		return "<%d>{{.TimeStamp}} %s %s[%s]: %s\n"
	}

	// Using RFC 5424 (IETF) format
	//              TIMESTAMP
	// <PRI>VERSION DATE TIME HOSTNAME APPNAME PID MSGID SD MESSAGE
	return "<%d>%d {{.TimeStamp}} %s %s %s %s %s %s\n"
}

// Determine the protocol to use to send records
// udp | tcp | tls
func setProtocol(op Operation) (string, error) {
	proto := `tcp`
	if op.Udp && op.Tls {
		return ``, fmt.Errorf("-udp and -tls not permitted")
	}

	if op.Udp {
		return `udp`, nil
	}

	if op.Tls {
		return `tls`, nil
	}

	return proto, nil
}

// The user can specify the count of messages to send. It just can't be < 0
// There is no upper limit. User wants to specify 2^63-1 on a 64 bit machine?
// Go for it
func validateCount(count int) error {
	if count <= 0 {
		return fmt.Errorf("-count must be > 0")
	}

	return nil
}

// Generate a random string of A-Z chars of len characters
func randomString(len int) *string {
	bytes := make([]byte, len)
	for i := range len {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	s := string(bytes)
	return &s
}

// Currently we are setting the syslog MESSAGE to a random string. With the wrapper
// over randomString() we can modify this to carry specific MESSAGES later. e.g.
// user points to a file or a pipe
func setMessage(length int) *string {
	return randomString(length)
}

// If the default priority is local0.info. The user would change the priority with
// '-pri facility.severity'. Here we convert the string based priority to an int
// for use in the Record
func setPriority(priority string) (int, error) {
	n, err := PriValAtoi(priority)
	if err != nil {
		return -1, err
	}

	return n, nil
}

// Go will try to give you the -w you desire, but if -w > GOMAXPROCS, you will have some go
// routines sitting in wait states
func checkConcurrency(w int) {
	cpu := runtime.NumCPU()
	var n int

	if w > cpu {
		fmt.Fprintf(os.Stderr, "%%warn: requested worker count -w=%d > NumCPU=%d\n", w, cpu)
		n = runtime.GOMAXPROCS(w)
		if w > n {
			fmt.Fprintf(os.Stderr, "%%note: GOMAXPROCS returned %d\n", n)
		}
	}
}

// Return the format to use for timestamps
func setTimeFormat(bsd bool) string {
	if bsd {
		return `Jan _2 15:04:05`
	}

	return `2006-01-02T15:04:05.000Z`
}

// Hydrate the string to be used as a template
func setTemplate(bsd bool, format string, r Record) string {
	if bsd {
		return fmt.Sprintf(format, r.Pri, r.Hostname, r.AppName, r.Pid, *r.Msg)
	}

	return fmt.Sprintf(format, r.Pri, r.Version, r.Hostname, r.AppName, r.Pid, r.MsgId, r.Sd, *r.Msg)
}

func milliSecondFormat(d time.Duration) string {
	totalMs := d.Milliseconds()
	day := totalMs / 86400000
	totalMs %= 86400000
	hours := totalMs / 3600000
	totalMs %= 3600000
	minutes := totalMs / 60000
	totalMs %= 60000
	seconds := totalMs / 1000
	totalMs %= 1000
	ms := totalMs
	return fmt.Sprintf("%02dd %02d:%02d:%02d.%03d", day, hours, minutes, seconds, ms)
}

func microSecondFormat(d time.Duration) string {
	totalUs := d.Microseconds()
	day := totalUs / 86400000000
	totalUs %= 86400000000
	hours := totalUs / 3600000000
	totalUs %= 3600000000
	minutes := totalUs / 60000000
	totalUs %= 60000000
	seconds := totalUs / 1000000
	totalUs %= 1000000
	ms := totalUs
	return fmt.Sprintf("%02dd %02d:%02d:%02d.%06d", day, hours, minutes, seconds, ms)
}

// SDG

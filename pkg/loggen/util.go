package loggen

import (
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

const (
	Success = iota
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

// Validate the supplied port is within range
func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	if p < 1 || p > 65535 {
		return fmt.Errorf("%%err: port value '%s' out of range", port)
	}

	return nil
}

func isValidIP(s string) bool {
	addr, err := netip.ParseAddr(s)
	return err == nil && addr.IsValid()
}

// Validate the destination. Could be a hostname, IP, stdout or file
func validateWDst(dst string) (string, error) {
	l := len(dst)
	if l == 0 {
		return ``, fmt.Errorf("%%err: -dst required")
	}

	if dst == `stdout` {
		return `stdout`, nil
	}

	// do we have a valid IP?
	if isValidIP(dst) {
		return `ip`, nil
	}

	// do we have a file?
	slash := regexp.MustCompile(`\/`)
	if slash.MatchString(dst) {
		return `file`, nil
	}

	// do we have a hostname?
	dot := regexp.MustCompile(`\.`)
	if l > 63 && !dot.MatchString(dst) { // non-fqdn hostname may not be > 63 char
		return ``, fmt.Errorf("%%err: dst hostname name exceeds maximum characters (63)")
	}

	if l > 253 && dot.MatchString(dst) { // an fqdn may not exceed 253 char
		return ``, fmt.Errorf("%%err: dst fqdn length exceeded. max = 253")
	}

	_, err := net.LookupIP(dst)
	if err != nil {
		return ``, fmt.Errorf("%%err: dns lookup for -dst '%s' failed", dst)
	}

	return `name`, nil
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

// Go will try to give you the -gr you desire, but if -gr > GOMAXPROCS, you will have some go
// routines sitting in wait states
func checkConcurrency(gr int) {
	cpu := runtime.NumCPU()
	var n int

	if gr > cpu {
		fmt.Fprintf(os.Stderr, "%%warn: requested go routine count -gr=%d > NumCPU=%d\n", gr, cpu)
		n = runtime.GOMAXPROCS(gr)
		if gr > n {
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

package loggen

import (
	"fmt"
	"strings"
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
		"emerg":    0, // Emergency: system is unusable
		"alert":    1, // Alert: action must be taken immediately
		"critical": 2, // Critical: critical conditions
		"error":    3, // Error: error conditions
		"warn":     4, // Warning: warning conditions
		"notice":   5, // Notice: normal but significant condition
		"info":     6, // Informational: informational messages
		"debug":    7, // Debug: debug-level messages
	}
)

// Receiving priorit as a string (e.g., 'local0.info') convert it to an integer
func PriValAtoi(priority string) (int, error) {
	sl := strings.Split(priority, ".")
	f := strings.ToLower(sl[0])
	s := strings.ToLower(sl[1])
	facility, facility_exists := Facility[f]
	severity, severity_exists := Severity[s]

	if facility_exists && severity_exists {
		// RFC5424 6.2.1 p11
		return (facility << 3) + severity, nil
	}

	return -1, fmt.Errorf("'%s' not supported", priority)
}

func invertMap(m map[string]int) map[int]string {
	im := make(map[int]string, len(m))
	for k, v := range m {
		im[v] = k
	}
	return im
}

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

package syslog

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
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
		"audit":    13,
		"alert":    14,
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

type RFC3164 struct {
	Format   *string
	Hostname string
	PRI      int
	PID      bool
	RFC3339  bool
	Tag      string
	Message  *string
}

type RFC5424 struct {
	Format   *string
	Version  int
	PRI      int
	Hostname string
	AppName  string
	ProcID   bool
	MsgID    string
	Sd       string
	Message  *string
}

// Set the syslog priority
func SetPriority(priority string) (int, error) {
	// priority = facility * 8 + severity
	s := strings.Split(priority, ".")
	facility, ok_facility := Facility[s[0]]
	severity, ok_severity := Severity[s[1]]

	if ok_facility && ok_severity {
		return facility*8 + severity, nil
	}

	return 0, fmt.Errorf("priority of '%s' is not supported", priority)
}

// Set the BSD syslog format string => <PRI>TIMESTAMP HOSTNAME TAG([\d+])?: MESSAGE
// To use the format string, only TIMESTAMP, and MESSAGE will need to be supplied
func SetBSDRecordFormat(r *RFC3164) {
	var (
		s    string
		_tag string
	)

	if r.PID {
		_tag = fmt.Sprintf("%s[%d]", r.Tag, os.Getpid())
	} else {
		_tag = r.Tag
	}

	s = fmt.Sprintf("<%d>%%s %s %s: %%s\n", r.PRI, r.Hostname, _tag)

	r.Format = &s
}

// Set the IETF syslog format string => <PRI>VERSION TIMESTAMP HOSTNAME APPNAME PROCID MSGID SD MESSAGE
// To use the format string, only TIMESTAMP, and MESSAGE will need to be supplied
func SetIETFRecordFormat(r *RFC5424) {
	var (
		s       string
		_procid string
	)

	if r.ProcID {
		_procid = fmt.Sprintf("%d", os.Getpid())
	} else {
		_procid = "-"
	}

	s = fmt.Sprintf("<%d>%d %%s %s %s %s %s %s %%s\n", r.PRI, r.Version, r.Hostname, r.AppName, _procid, r.MsgID, r.Sd)

	r.Format = &s
}

type kvPair struct {
	Key   string
	Value int
}

type kvPairList []kvPair

func (p kvPairList) Len() int           { return len(p) }
func (p kvPairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p kvPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func FacilityList() {
	kvp := make(kvPairList, len(Facility))
	i := 0
	for k, v := range Facility {
		kvp[i] = kvPair{k, v}
		i++
	}

	//descending -> sort.Sort(sort.Reverse(kvp))
	sort.Sort(kvp)
	for _, v := range kvp {
		fmt.Printf("%10s   %d\n", v.Key, v.Value)
	}
}

func SeverityList() {
	kvp := make(kvPairList, len(Severity))
	i := 0
	for k, v := range Severity {
		kvp[i] = kvPair{k, v}
		i++
	}

	//descending -> sort.Sort(sort.Reverse(kvp))
	sort.Sort(kvp)
	for _, v := range kvp {
		fmt.Printf("%10s   %d\n", v.Key, v.Value)
	}
}

//SDG

package loggen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/rpcox/pkg/exit"
)

type ClientFunc func(int, Loggen, *Interrupt, *sync.WaitGroup)

// Report errors to console
func ReportWriteError(err error) {
	if errors.Is(err, syscall.EPIPE) {
		fmt.Fprintf(os.Stderr, "%%err: EPIPE: broken pipe\n")
		return
	}
	if errors.Is(err, syscall.ECONNRESET) {
		fmt.Fprintf(os.Stderr, "%%err: ECONNRESET: connection reset by peer\n")
		return
	}

	if errors.Is(err, syscall.EBADF) {
		fmt.Fprintf(os.Stderr, "%%err: EBADF: bad file descriptorr\n")
		return
	}
	if errors.Is(err, syscall.EINTR) {
		fmt.Fprintf(os.Stderr, "%%err: EINTR: write interrupted\n")
		return
	}
	if errors.Is(err, net.ErrClosed) {
		fmt.Fprintf(os.Stderr, "%%err: write attempt on closed socket\n")
		return
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		fmt.Fprintf(os.Stderr, "%%err: %s failed on %s\n", opErr.Op, opErr.Net)

		if opErr.Timeout() {
			fmt.Fprintf(os.Stderr, "%%err: %s write deadline exceeded", opErr.Net)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "%%err: %v\n", err)
}

func udpClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	worker := fmt.Sprintf(`udp-worker-%02d`, i)
	network := `udp`
	Carp(fmt.Sprintf("%s: starting", worker))

	port := strconv.Itoa(lg.opt.Destination.Port)
	dst := lg.opt.Destination.Dst + `:` + port
	addr, err := net.ResolveUDPAddr(network, dst)
	exit.If(err != nil, err, ErrResolveAddrFail)
	conn, err := net.DialUDP(network, nil, addr)
	exit.If(err != nil, err, ErrUdpClientFail)
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer conn.Close()
		defer intr.mu.Unlock()

		<-sigChan
		intr.mu.Lock()
		if !intr.yes {
			intr.yes = true
			intr.at = time.Now().UTC().Format(timeFormat)
		}
	}()

	var S Stats
	tmpl := template.Must(template.New("udp-record").Parse(lg.RecordTmpl))
	start := time.Now()
	if lg.opt.Operation.timeDuration == 0 {
		S = udpByCount(conn, lg.opt, tmpl, lg.TimeFormat)
	} else {
		S = udpWithContext(conn, lg.opt, tmpl, lg.TimeFormat)
	}

	S.Worker = worker
	S.Duration = time.Since(start)
	lg.statsChan <- S
}

// Send UDP records down range by count
func udpByCount(conn *net.UDPConn, opt *Options, tmpl *template.Template, format string) Stats {
	var buf bytes.Buffer
	var n int
	var stamp Ts
	var totalBytes uint64
	var S Stats

	for range opt.Operation.Count {
		S.Count++
		stamp.TimeStamp = time.Now().Format(format)
		err := tmpl.Execute(&buf, stamp)
		exit.If(err != nil, err, ErrTemplate)
		n, err = conn.Write(buf.Bytes())
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			} else {
				ReportWriteError(err)
				os.Exit(ErrNetworkWrite)
			}
		}

		totalBytes += uint64(n)
		buf.Reset()
	}

	S.Bytes = totalBytes
	return S
}

// Send UDP records down range with timeout
func udpWithContext(conn *net.UDPConn, opt *Options, tmpl *template.Template, format string) Stats {
	var buf bytes.Buffer
	var n int
	var stamp Ts
	var totalBytes uint64
	var S Stats

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opt.Operation.timeDuration.Seconds())*time.Second)
	defer cancel()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		default:
			S.Count++
			stamp.TimeStamp = time.Now().Format(format)
			err := tmpl.Execute(&buf, stamp)
			exit.If(err != nil, err, ErrTemplate)
			n, err = conn.Write(buf.Bytes())
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break Loop
				} else {
					ReportWriteError(err)
					os.Exit(ErrNetworkWrite)
				}
			}
			totalBytes += uint64(n)
			buf.Reset()
		}
	}

	S.Bytes = totalBytes
	return S
}

func tcpClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	worker := fmt.Sprintf(`tcp-worker-%02d`, i)
	network := `tcp`
	Carp(fmt.Sprintf("%s: starting", worker))

	port := strconv.Itoa(lg.opt.Destination.Port)
	dst := lg.opt.Destination.Dst + `:` + port
	addr, err := net.ResolveTCPAddr(network, dst)
	exit.If(err != nil, err, ErrResolveAddrFail)
	conn, err := net.DialTCP(network, nil, addr)
	exit.If(err != nil, err, ErrTcpClientFail)
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer conn.Close()
		defer intr.mu.Unlock()

		<-sigChan
		intr.mu.Lock()
		if !intr.yes {
			intr.yes = true
			intr.at = time.Now().UTC().Format(timeFormat)
		}
	}()

	var S Stats
	tmpl := template.Must(template.New("tcp-record").Parse(lg.RecordTmpl))
	start := time.Now()
	if lg.opt.Operation.timeDuration == 0 {
		S = tcpByCount(conn, lg.opt, tmpl, lg.TimeFormat)
	} else {
		S = tcpWithContext(conn, lg.opt, tmpl, lg.TimeFormat)
	}

	S.Worker = worker
	S.Duration = time.Since(start)
	lg.statsChan <- S
}

// Send TCP records down range by count
func tcpByCount(conn *net.TCPConn, opt *Options, tmpl *template.Template, format string) Stats {
	var buf bytes.Buffer
	var n int
	var stamp Ts
	var totalBytes uint64
	var S Stats

	for range opt.Operation.Count {
		S.Count++
		stamp.TimeStamp = time.Now().Format(format)
		err := tmpl.Execute(&buf, stamp)
		exit.If(err != nil, err, ErrTemplate)
		n, err = conn.Write(buf.Bytes())
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			} else {
				ReportWriteError(err)
				os.Exit(ErrNetworkWrite)
			}
		}

		totalBytes += uint64(n)
		buf.Reset()
	}

	S.Bytes = totalBytes
	return S
}

// Send TCP records down range with timeout
func tcpWithContext(conn *net.TCPConn, opt *Options, tmpl *template.Template, format string) Stats {
	var buf bytes.Buffer
	var n int
	var stamp Ts
	var totalBytes uint64
	var S Stats

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opt.Operation.timeDuration.Seconds())*time.Second)
	defer cancel()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		default:
			S.Count++
			stamp.TimeStamp = time.Now().Format(format)
			err := tmpl.Execute(&buf, stamp)
			exit.If(err != nil, err, ErrTemplate)
			n, err = conn.Write(buf.Bytes())
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break Loop
				} else {
					ReportWriteError(err)
					os.Exit(ErrNetworkWrite)
				}
			}
			totalBytes += uint64(n)
			buf.Reset()
		}
	}

	S.Bytes = totalBytes
	return S
}

func tlsClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	Carp(fmt.Sprintf("tls-worker[%d]: starting", i))
	port := strconv.Itoa(lg.opt.Destination.Port)
	dst := lg.opt.Destination.Dst + `:` + port
	fmt.Println(dst)
}

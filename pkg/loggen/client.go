package loggen

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/rpcox/pkg/exit"
)

type clientFunc func(int, Loggen, *Interrupt, *sync.WaitGroup)

func udpClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	worker := fmt.Sprintf(`udp-worker-%02d`, i)
	network := `udp`
	Carp(fmt.Sprintf("%s: starting", worker))

	dst := lg.opt.Destination.Dst + `:` + lg.opt.Destination.Port
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

	var buf bytes.Buffer
	var n int
	var totalBytes uint64
	var stamp Ts
	var S Stats
	S.Worker = worker

	tmpl := template.Must(template.New("udp-record").Parse(lg.RecordTmpl))
	start := time.Now()
	for range lg.opt.Operation.Count {
		S.Count++
		stamp.TimeStamp = time.Now().Format(lg.TimeFormat)
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

	S.Duration = time.Since(start)
	S.Bytes = totalBytes
	lg.statsChan <- S
}

func tcpClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	worker := fmt.Sprintf(`tcp-worker-%02d`, i)
	network := `tcp`
	Carp(fmt.Sprintf("%s: starting", worker))

	dst := lg.opt.Destination.Dst + `:` + lg.opt.Destination.Port
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

	var buf bytes.Buffer
	var n int
	var totalBytes uint64
	var stamp Ts
	var S Stats
	S.Worker = worker

	tmpl := template.Must(template.New("tcp-record").Parse(lg.RecordTmpl))
	start := time.Now()
	for range lg.opt.Operation.Count {
		S.Count++
		stamp.TimeStamp = time.Now().Format(lg.TimeFormat)
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

	S.Duration = time.Since(start)
	S.Bytes = totalBytes
	lg.statsChan <- S
}

func tlsClient(i int, lg Loggen, intr *Interrupt, wg *sync.WaitGroup) {
	defer wg.Done()

	Carp(fmt.Sprintf("tls-worker[%d]: starting", i))
	dst := lg.opt.Destination.Dst + `:` + lg.opt.Destination.Port
	fmt.Println(dst)
}

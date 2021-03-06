package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("main")

var opts struct {
	TargetAddress string        `short:"t" long:"target" description:"The target address in host:port form" required:"true"`
	Port          int           `short:"p" long:"port" description:"The port Slowpoke should listen for connections on" required:"true"`
	Verbose       []bool        `short:"v" long:"verbose" description:"Log verbosity level. -v or -vv"`
	Latency       time.Duration `short:"l" long:"latency" default:"0ms" description:"The duration of latency to apply to data packets, specified as a number and unit. E.g. 15ms or 2s. Supported units are 'us', 'ms', 's', 'm' and 'h'"`
	BufferSize    int           `short:"b" long:"buffer" default:"1500" description:"The size of the transfer buffer in bytes. Latency is applied between each buffer flush. Therefore total latency applied is equal to '(totalDataTransferred/bufferSize) * latency'"`
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	configureLogger()
}

func configureLogger() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logFormat := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} [%{level:.3s}]%{color:reset} - %{message}`)
	logger := logging.AddModuleLevel(logging.NewBackendFormatter(logBackend, logFormat))

	if len(opts.Verbose) == 0 {
		logger.SetLevel(logging.WARNING, "")
	} else if len(opts.Verbose) == 1 {
		logger.SetLevel(logging.INFO, "")
	} else {
		logger.SetLevel(logging.DEBUG, "")
	}

	logging.SetBackend(logger)
}

func main() {
	log.Infof("Proxying between :%d and %s with %s of latency", opts.Port, opts.TargetAddress, opts.Latency)
	log.Debugf("Transfer buffer size set to %d bytes", opts.BufferSize)

	listener := getListener(opts.Port)
	targetAddr := resolveTarget(opts.TargetAddress)
	waitForClients(listener, targetAddr)
}

func getListener(port int) net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Errorf("Failed to start listening on port %d:\n%v", port, err)
		os.Exit(1)
	}
	log.Debugf("Waiting for connections on port %d", port)

	return listener
}

func resolveTarget(targetAddress string) *net.TCPAddr {
	tcpAddr, err := net.ResolveTCPAddr("tcp", opts.TargetAddress)
	if err != nil {
		log.Errorf("Failed to resolve target address %s:\n%v", opts.TargetAddress, err)
		os.Exit(1)
	}
	return tcpAddr
}

func waitForClients(listener net.Listener, targetAddr *net.TCPAddr) {
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept connection:\n%v", err)
			break
		}
		log.Infof("Accepted connection from client %v\n", client.RemoteAddr())

		s := NewSlowpoke(client, targetAddr, opts.Latency, opts.BufferSize, log)

		go s.StartTransfer()
	}
}

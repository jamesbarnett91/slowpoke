package main

import (
	"io"
	"net"
	"time"

	"github.com/op/go-logging"
)

type Slowpoke struct {
	conn       net.Conn
	targetAddr *net.TCPAddr
	latency    time.Duration
	bufferSize int
	isClosed   bool
	close      chan bool
	logger     *logging.Logger
}

func NewSlowpoke(conn net.Conn, targetAddr *net.TCPAddr, latency time.Duration, bufferSize int, logger *logging.Logger) *Slowpoke {
	return &Slowpoke{
		conn:       conn,
		targetAddr: targetAddr,
		latency:    latency,
		bufferSize: bufferSize,
		isClosed:   false,
		close:      make(chan bool),
		logger:     logger,
	}
}

func (s *Slowpoke) StartTransfer() {
	defer s.conn.Close()
	target, err := net.DialTCP("tcp", nil, s.targetAddr)
	if err != nil {
		s.logger.Errorf("Failed to connect to target address %s:\n%v", s.targetAddr, err)
		return
	}
	defer target.Close()
	s.logger.Debugf("Established connection to %s", target.RemoteAddr())

	go s.transferWithLatency(s.conn, target)
	go s.transferWithLatency(target, s.conn)

	<-s.close

	s.logger.Infof("Connection between client %s and target %s closed", s.conn.RemoteAddr(), target.RemoteAddr())
}

func (s *Slowpoke) createBuffer() []byte {
	return make([]byte, s.bufferSize)
}

func (s *Slowpoke) transferWithLatency(source net.Conn, target net.Conn) {
	byteBuffer := s.createBuffer()

	for {

		bytesRead, readError := source.Read(byteBuffer)

		if bytesRead > 0 {

			s.logger.Debugf("Transferring %d bytes from %s to %s with %s added latency", bytesRead, source.RemoteAddr(), target.RemoteAddr(), s.latency)

			if s.latency != 0 {
				time.Sleep(s.latency)
			}

			bytesWritten, writeError := target.Write(byteBuffer[0:bytesRead])

			if writeError != nil {
				s.handleError("Error during write: %v", writeError)
				break
			}
			if bytesRead != bytesWritten {
				s.logger.Warningf("Read %d bytes but could only write %d bytes", bytesRead, bytesWritten)
			}
		}
		if readError != nil {
			s.handleError("Error during read: %v", readError)
			break
		}
	}

}

func (s *Slowpoke) handleError(msg string, err error) {
	if s.isClosed {
		// One of the send/receive streams was already closed. Nothing to do.
		return
	}

	s.isClosed = true
	s.close <- true

	if err == io.EOF {
		// EOF is expected and not really an error
		s.logger.Debug("Received EOF")
	} else {
		s.logger.Errorf(msg, err)
	}
}

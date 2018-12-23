package slowpoke

import (
	"io"
	"net"
	"time"

	"github.com/op/go-logging"
)

type Slowpoke struct {
	conn       net.Conn
	targetAddr string
	latency    time.Duration
	isClosed   bool
	close      chan bool
	logger     *logging.Logger
}

func New(conn net.Conn, targetAddr string, latency time.Duration, logger *logging.Logger) *Slowpoke {
	return &Slowpoke{
		conn:       conn,
		targetAddr: targetAddr,
		latency:    latency,
		isClosed:   false,
		close:      make(chan bool),
		logger:     logger,
	}
}

func (s *Slowpoke) StartTransfer() {
	defer s.conn.Close()
	target, err := net.Dial("tcp", s.targetAddr)
	if err != nil {
		// TODO validate target addr before this point
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

func createBuffer() []byte {
	// TODO configurable
	return make([]byte, 1500)
}

func (s *Slowpoke) transferWithLatency(source net.Conn, target net.Conn) {
	byteBuffer := createBuffer()

	var transferDirection string
	// If the data source is the client then we are sending
	if source == s.conn {
		transferDirection = "%d bytes sent"
	} else {
		transferDirection = "%d bytes received"
	}

	for {

		bytesRead, readError := source.Read(byteBuffer)

		if bytesRead > 0 {

			if s.latency != 0 {
				s.logger.Debugf(transferDirection+" with latency of %s", bytesRead, s.latency)
				time.Sleep(s.latency)
			} else {
				s.logger.Debugf(transferDirection, bytesRead)
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

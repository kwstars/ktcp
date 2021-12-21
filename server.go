package ktcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kwstars/ktcp/internal/ksync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/kwstars/ktcp/encoding"
	"github.com/kwstars/ktcp/encoding/proto"
	"github.com/kwstars/ktcp/packing"
)

// Byte unit helpers.
const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
)

// Handler Server
type Handler interface {
	CallBack
	OnConnect(s *Session)
}

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// ReadTimeout with server timeout.
func ReadTimeout(readTimeout time.Duration) ServerOption {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

// WriteTimeout with server timeout.
func WriteTimeout(writeTimeout time.Duration) ServerOption {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}

// Logger with server logger.
func Logger(logger log.Logger) ServerOption {
	return func(s *Server) {
		s.log = log.NewHelper(logger)
	}
}

// Middleware with service middleware option.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.ms = m
	}
}

// ErrServerStopped is returned when server stopped.
var ErrServerStopped = errors.New("ktcp: the server has been stopped")

// ServerOption is an HTTP server option.
type ServerOption func(*Server)

// Server is a server for TCP connections.
type Server struct {
	socketReadBufferSize  int
	socketWriteBufferSize int
	reqQueueSize          int
	respQueueSize         int
	writeAttemptTimes     int
	readTimeout           time.Duration
	writeTimeout          time.Duration
	network               string
	address               string
	Listener              net.Listener
	Packer                packing.Packer // Packer is the message packer, will be passed to session.
	Codec                 encoding.Codec // Codec is the message codec, will be passed to session.
	callback              Handler
	quit                  *ksync.Event
	log                   *log.Helper
	ms                    []middleware.Middleware
	serveWG               sync.WaitGroup
	pool                  *sync.Pool
	sessions              sync.Map
}

// NewServer creates an TCP server by options.
func NewServer(handler Handler, opts ...ServerOption) *Server {
	srv := &Server{
		socketReadBufferSize:  1 * MB,
		socketWriteBufferSize: 1 * MB,
		reqQueueSize:          1024,
		respQueueSize:         1024,
		writeAttemptTimes:     1,
		readTimeout:           3 * time.Second,
		writeTimeout:          3 * time.Second,
		network:               "tcp",
		address:               ":9090",
		Packer:                packing.NewDefaultPacker(),
		Codec:                 proto.New(),
		callback:              handler,
		serveWG:               sync.WaitGroup{},
		log:                   log.NewHelper(log.DefaultLogger),
		pool:                  &sync.Pool{New: func() interface{} { return NewContext() }},
		quit:                  ksync.NewEvent(),
	}

	logger := log.NewHelper(log.DefaultLogger)
	srv.log = logger

	for _, o := range opts {
		o(srv)
	}

	return srv
}

// Serve the TCP server
func (s *Server) Serve() error {
	address, err := net.ResolveTCPAddr(s.network, s.address)
	if err != nil {
		return err
	}

	s.log.Infof("start tcp server at %s", address.String())
	lis, err := net.ListenTCP(s.network, address)
	if err != nil {
		return err
	}

	s.Listener = lis

	var tempDelay time.Duration

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if ne, ok := err.(interface{ Temporary() bool }); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				s.log.Errorf("http: Accept error: %v; retrying in %v", err, tempDelay)
				timer := time.NewTimer(tempDelay)
				select {
				case <-timer.C:
				case <-s.quit.Done():
					timer.Stop()
					return nil
				}
				continue
			}
			if s.quit.HasFired() {
				return nil
			}
			return err
		}

		tempDelay = 0

		if s.socketReadBufferSize > 0 {
			if err = conn.(*net.TCPConn).SetReadBuffer(s.socketReadBufferSize); err != nil {
				return fmt.Errorf("conn set read buffer err: %s", err)
			}
		}
		if s.socketWriteBufferSize > 0 {
			if err = conn.(*net.TCPConn).SetWriteBuffer(s.socketWriteBufferSize); err != nil {
				return fmt.Errorf("conn set write buffer err: %s", err)
			}
		}

		s.serveWG.Add(1)
		go func() {
			s.handleRawConn(conn)
			s.serveWG.Done()
		}()
	}
}

// handleRawConn handles the connection
func (s *Server) handleRawConn(conn net.Conn) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	sess := newSession(conn, s, cancelFunc)

	s.sessions.Store(sess.ID(), sess)
	defer func() {
		s.removeSession(sess)
	}()

	s.callback.OnConnect(sess)

	if err := sess.readInbound(ctx); err != nil {
		s.log.Errorf("session read inbound err: %s", err)
	}

	s.callback.OnClose(sess)
}

// Stop the TCP server.
func (s *Server) Stop(ctx context.Context) (err error) {
	s.quit.Fire()

	// close the listener
	if err := s.Listener.Close(); err != nil {
		s.log.Errorf("close listener err: %s", err)
	}

	// close sessions
	s.sessions.Range(func(k, v interface{}) bool {
		v.(*Session).Close()
		return true
	})

	return
}

func (s *Server) removeSession(sess *Session) {
	s.sessions.Delete(sess.ID())
	sess.Close()
}

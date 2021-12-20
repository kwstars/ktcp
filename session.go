package ktcp

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/kwstars/ktcp/encoding"
	"github.com/kwstars/ktcp/message"
	"github.com/kwstars/ktcp/packing"
	"github.com/kwstars/ktcp/sync/atomic"
	"github.com/segmentio/ksuid"
	"net"
	"time"
)

const (
	tempErrDelay = time.Millisecond * 5
)

// ErrSessionClosed is returned when session stopped.
var ErrSessionClosed = fmt.Errorf("session closed")

type CallBack interface {
	OnMessage(c Context)
	OnClose(c *Session)
}

// Session is a server side network connection.
type Session struct {
	connected  atomic.Bool
	id         string                // session's ID. it's a UUID
	conn       net.Conn              // tcp connection
	respQueue  chan Context          // response queue channel, pushed in SendResp() and popped in writeOutbound()
	reqQueue   chan *message.Message // request queue channel, pushed in readInbound() and popped in Handle()
	packer     packing.Packer        // to pack and unpack message
	codec      encoding.Codec        // encode/decode message data
	callback   CallBack
	log        *log.Helper
	cancelFunc context.CancelFunc
}

func (s *Session) Codec() encoding.Codec {
	return s.codec
}

// newSession creates a new session.
func newSession(conn net.Conn, s *Server, cancelFunc context.CancelFunc) (sess *Session) {
	sess = &Session{
		conn:       conn,
		cancelFunc: cancelFunc,
		id:         ksuid.New().String(),
		reqQueue:   make(chan *message.Message, s.reqQueueSize),
		respQueue:  make(chan Context, s.respQueueSize),
		packer:     s.Packer,
		codec:      s.Codec,
		callback:   s.callback,
		log:        s.log,
	}

	sess.connected.SetTrue()

	return
}

// ID returns the session's ID.
func (s *Session) ID() string {
	return s.id
}

// Send pushes response message entry to respQueue.
func (s *Session) Send(ctx Context) (err error) {
	if s.connected.IsSet() {
		s.respQueue <- ctx
		return nil
	}
	return ErrSessionClosed
}

// Close closes the session, but doesn't close the connection.
func (s *Session) Close() {
	s.connected.SetFalse()
	s.cancelFunc()
	if err := s.conn.Close(); err != nil {
		s.log.Errorf("connection close err: %s", err)
	}
}

// readInbound reads message packet from connection in a loop.
func (s *Session) readInbound(ctx context.Context, doneChan chan<- struct{}, router *Router, timeout time.Duration) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info("readInbound", ctx.Err())
			return
		default:
			if timeout > 0 {
				if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
					s.log.Errorf("session %s set read deadline err: %s", s.id, err)
					doneChan <- struct{}{}
					return
				}
			}

			reqMsg, err := s.packer.Unpack(s.conn)
			if err != nil {
				s.log.Errorf("session %s unpack inbound packet err: %s", s.id, err)
				doneChan <- struct{}{}
				return
			}

			if reqMsg == nil {
				continue
			}

			// handle request
			go func() {
				routerCtx := router.pool.Get().(Context)
				routerCtx.Reset(s, reqMsg, nil)
				s.callback.OnMessage(routerCtx)
				router.pool.Put(routerCtx)
			}()
		}
	}
}

// writeOutbound fetches message from respQueue channel and writes to TCP connection in a loop.
func (s *Session) writeOutbound(ctx context.Context, doneChan chan<- struct{}, writeTimeout time.Duration, attemptTimes int) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info("writeOutbound: ", ctx.Err())
			return
		default:
			rCtx, ok := <-s.respQueue
			if !ok {
				s.log.Errorf("session %s respQueue closed", s.id)
				doneChan <- struct{}{}
				return
			}

			outboundMsg, err := s.packResponse(rCtx)
			if err != nil {
				s.log.Errorf("session %s pack outbound message err: %s", s.id, err)
				continue
			}
			if outboundMsg == nil {
				s.log.Errorf("session %s out message is nil", s.id)
				continue
			}
			if writeTimeout > 0 {
				if err = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
					s.log.Errorf("session %s set write deadline err: %s", s.id, err)
					doneChan <- struct{}{}
					return
				}
			}
			if err = s.attemptConnWrite(outboundMsg, attemptTimes); err != nil {
				s.log.Errorf("session %s conn write err: %s", s.id, err)
				doneChan <- struct{}{}
				return
			}
		}
	}
}

func (s *Session) attemptConnWrite(outboundMsg []byte, attemptTimes int) (err error) {
	for i := 0; i < attemptTimes; i++ {
		time.Sleep(tempErrDelay * time.Duration(i))
		_, err = s.conn.Write(outboundMsg)

		// breaks if err is not nil or it's the last attempt.
		if err == nil || i == attemptTimes-1 {
			break
		}

		// check if err is `net.Error`
		ne, ok := err.(net.Error)
		if !ok {
			break
		}
		if ne.Timeout() {
			break
		}
		if ne.Temporary() {
			s.log.Errorf("session %s conn write err: %s; retrying in %s", s.id, err, tempErrDelay*time.Duration(i+1))
			continue
		}
		break // if err is not temporary, break the loop.
	}
	return
}

func (s *Session) packResponse(ctx Context) ([]byte, error) {
	if ctx.Response() == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.Response())
}

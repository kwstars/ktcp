package ktcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kwstars/ktcp/sync/atomic"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/kwstars/ktcp/encoding"
	"github.com/kwstars/ktcp/message"
	"github.com/kwstars/ktcp/packing"
	"github.com/segmentio/ksuid"
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
	connected         atomic.Bool
	writeAttemptTimes int
	id                string                // session's ID. it's a UUID
	conn              net.Conn              // tcp connection
	respQueue         chan Context          // response queue channel, pushed in SendResp() and popped in writeOutbound()
	reqQueue          chan *message.Message // request queue channel, pushed in readInbound() and popped in Handle()
	packer            packing.Packer        // to pack and unpack message
	codec             encoding.Codec        // encode/decode message data
	callback          CallBack
	log               *log.Helper
	cancelFunc        context.CancelFunc
	pool              *sync.Pool
}

func (s *Session) Codec() encoding.Codec {
	return s.codec
}

// newSession creates a new session.
func newSession(conn net.Conn, s *Server, cancelFunc context.CancelFunc) (sess *Session) {
	sess = &Session{
		conn:              conn,
		cancelFunc:        cancelFunc,
		id:                ksuid.New().String(),
		reqQueue:          make(chan *message.Message, s.reqQueueSize),
		respQueue:         make(chan Context, s.respQueueSize),
		writeAttemptTimes: s.writeAttemptTimes,
		packer:            s.Packer,
		codec:             s.Codec,
		callback:          s.callback,
		log:               s.log,
		pool:              s.pool,
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
	outboundMsg, err := s.packResponse(ctx)
	if err != nil {
		return fmt.Errorf("session %s pack outbound message err: %s", s.id, err)
	}

	if outboundMsg == nil {
		return fmt.Errorf("session %s out message is nil", s.id)
	}

	if err = s.attemptConnWrite(outboundMsg, s.writeAttemptTimes); err != nil {
		return fmt.Errorf("session %s conn write err: %s", s.id, err)
	}

	return
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
func (s *Session) readInbound(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			s.log.Info("readInbound", ctx.Err())
			return
		default:
			reqMsg, err := s.packer.Unpack(s.conn)
			if err != nil {
				return fmt.Errorf("session %s unpack inbound packet err: %s", s.id, err)
			}

			if reqMsg == nil {
				continue
			}

			go func(ctx context.Context) {
				routerCtx := s.pool.Get().(*routerCtx)
				routerCtx.Reset(s, reqMsg)
				s.callback.OnMessage(routerCtx)
				s.pool.Put(routerCtx)
			}(ctx)
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
		_, err = s.conn.Write(outboundMsg)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				s.log.Info("session %s write attempt %d temporary err: %s", s.id, i+1, err)
				time.Sleep(tempErrDelay * time.Duration(i))
				continue
			} else {
				return fmt.Errorf("type assertion: %v, session %s write attempt %d err: %s", ok, s.id, i+1, err)
			}
		}
		return
	}

	return fmt.Errorf("attemptConnWrite write failure threshold exceeded")
}

func (s *Session) packResponse(ctx Context) ([]byte, error) {
	if ctx.Response() == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.Response())
}

package ktcp

import (
	"context"
	"fmt"
	"time"

	"github.com/kwstars/ktcp/message"

	"github.com/go-kratos/kratos/v2/middleware"
)

var _ Context = (*routerCtx)(nil)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context
	GetSession() *Session
	ForwardHandler(c CallBack)
	GetReqMsg() *message.Message
	Bind(v interface{}) error
	Response() *message.Message
	Send(id uint32, flag uint16, resp interface{}) error
	Middleware(middleware.Handler) middleware.Handler
	Reset(sess *Session, reqMsg *message.Message)
}

type routerCtx struct {
	session *Session
	reqMsg  *message.Message
	respMsg *message.Message
}

func NewContext() *routerCtx {
	return &routerCtx{}
}

func (c *routerCtx) Deadline() (time.Time, bool) {
	return c.Deadline()
}

func (c *routerCtx) Done() <-chan struct{} {
	return c.Done()
}

func (c *routerCtx) Err() error {
	return c.Err()
}

func (c *routerCtx) Value(key interface{}) interface{} {
	return c.Value(key)
}

func (c *routerCtx) GetReqMsg() *message.Message {
	return c.reqMsg
}

func (c *routerCtx) ForwardHandler(callback CallBack) {
	c.session.callback = callback
}

func (c *routerCtx) Reset(sess *Session, reqMsg *message.Message) {
	c.session = sess
	c.reqMsg = reqMsg
	c.respMsg = nil
}

func (c *routerCtx) Middleware(h middleware.Handler) middleware.Handler {
	//TODO middleware

	return middleware.Chain()(h)
}

func (c *routerCtx) GetSession() *Session {
	return c.session
}

func (c *routerCtx) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Unmarshal(c.reqMsg.Data, v)
}

func (c *routerCtx) Response() *message.Message {
	return c.respMsg
}

func (c *routerCtx) Send(id uint32, flag uint16, data interface{}) error {

	codec := c.session.Codec()
	if codec == nil {
		return fmt.Errorf("message codec is nil")
	}
	dataRaw, err := codec.Marshal(data)
	if err != nil {
		return err
	}

	c.respMsg = &message.Message{
		ID:   id,
		Flag: flag,
		Data: dataRaw,
	}

	return c.session.Send(c)
}

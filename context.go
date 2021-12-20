package ktcp

import (
	"context"
	"fmt"
	"time"

	"github.com/kwstars/ktcp/message"

	"github.com/go-kratos/kratos/v2/middleware"
)

var _ Context = (*RouterCtx)(nil)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context
	GetSession() *Session
	ForwardHandler(c CallBack)
	GetRouter() *Router
	GetReqMsg() *message.Message
	Bind(v interface{}) error
	Response() *message.Message
	Send(id uint32, flag uint16, resp interface{}) error
	Middleware(middleware.Handler) middleware.Handler
	Reset(sess *Session, reqMsg *message.Message)
}

type RouterCtx struct {
	router  *Router
	session *Session
	reqMsg  *message.Message
	respMsg *message.Message
}

func (c *RouterCtx) Deadline() (time.Time, bool) {
	return c.Deadline()
}

func (c *RouterCtx) Done() <-chan struct{} {
	return c.Done()
}

func (c *RouterCtx) Err() error {
	return c.Err()
}

func (c *RouterCtx) Value(key interface{}) interface{} {
	return c.Value(key)
}

// NewRouterContext returns a new Context for the given request and response.
func NewRouterContext(r *Router) *RouterCtx {
	return &RouterCtx{
		router: r,
	}
}

func (c *RouterCtx) GetReqMsg() *message.Message {
	return c.reqMsg
}

func (c *RouterCtx) GetRespMsg() *message.Message {
	return c.respMsg
}

func (c *RouterCtx) GetRouter() *Router {
	return c.router
}

func (c *RouterCtx) ForwardHandler(callback CallBack) {
	c.session.callback = callback
}

func (c *RouterCtx) Reset(sess *Session, reqMsg *message.Message) {
	c.session = sess
	c.reqMsg = reqMsg
	c.respMsg = nil
}

func (c *RouterCtx) Middleware(h middleware.Handler) middleware.Handler {
	return middleware.Chain(c.router.srv.ms...)(h)
}

func (c *RouterCtx) GetSession() *Session {
	return c.session
}

func (c *RouterCtx) Bind(v interface{}) error {
	if c.session.Codec() == nil {
		return fmt.Errorf("message codec is nil")
	}
	return c.session.Codec().Unmarshal(c.reqMsg.Data, v)
}

func (c *RouterCtx) Response() *message.Message {
	return c.respMsg
}

func (c *RouterCtx) Send(id uint32, flag uint16, data interface{}) error {

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

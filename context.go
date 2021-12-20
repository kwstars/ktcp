package ktcp

import (
	"context"
	"fmt"
	"time"

	"github.com/kwstars/ktcp/message"

	"github.com/go-kratos/kratos/v2/middleware"
)

var _ Context = (*RouterCtx)(nil)

// NewRouterContext returns a new Context for the given request and response.
func NewRouterContext(r *Router) *RouterCtx {
	return &RouterCtx{
		router: r,
	}
}

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context

	// GetSession returns the current session.
	GetSession() *Session

	// SetSession sets session.
	SetSession(sess *Session) Context

	// ForwardHandler Change callback handler.
	ForwardHandler(c CallBack)

	// GetRouter returns the current router.
	GetRouter() *Router

	// GetReqMsg returns the current request message.
	GetReqMsg() *message.Message

	// GetRespMsg returns the current response message.
	GetRespMsg() *message.Message

	// Request returns request message Msg.
	Request() *message.Message

	// SetRequest encodes data with session's codec and sets request message Msg.
	SetRequest(id, data interface{}) error

	// MustSetRequest encodes data with session's codec and sets request message Msg.
	// panics on error.
	MustSetRequest(id, data interface{}) Context

	// SetRequestMessage sets request message Msg directly.
	SetRequestMessage(Msg *message.Message) Context

	// Bind decodes request message Msg to v.
	Bind(v interface{}) error

	// Response returns the response message Msg.
	Response() *message.Message

	// SetResponse encodes data with session's codec and sets response message Msg.
	MarshalResp(id uint32, flag uint16, data interface{}) error

	// MustSetResponse encodes data with session's codec and sets response message Msg.
	// panics on error.
	MustSetResponse(id, data interface{}) Context

	// SetResponseMessage sets response message Msg directly.
	SetResponseMessage(Msg *message.Message) Context

	// Send sends itself to current session.
	Send(id uint32, flag uint16, resp interface{}) error

	// SendTo sends itself to session.
	SendTo(session *Session)

	// Get returns key value from storage.
	Get(key string) (value interface{}, exists bool)

	// Set store key value into storage.
	Set(key string, value interface{})

	// Remove deletes the key from storage.
	Remove(key string)

	// Copy returns a copy of Context.
	Copy() Context

	Middleware(middleware.Handler) middleware.Handler

	Reset(sess *Session, reqMsg *message.Message, respMsg *message.Message)
}

type RouterCtx struct {
	router  *Router
	session *Session
	reqMsg  *message.Message
	respMsg *message.Message
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

func (c *RouterCtx) Reset(sess *Session, reqMsg *message.Message, respMsg *message.Message) {
	c.session = sess
	c.reqMsg = reqMsg
	c.respMsg = nil
}

func (c *RouterCtx) Middleware(h middleware.Handler) middleware.Handler {
	return middleware.Chain(c.router.srv.ms...)(h)
}

func (c *RouterCtx) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (c *RouterCtx) Done() <-chan struct{} {
	panic("implement me")
}

func (c *RouterCtx) Err() error {
	panic("implement me")
}

func (c *RouterCtx) Value(key interface{}) interface{} {
	panic("implement me")
}

func (c *RouterCtx) GetSession() *Session {
	return c.session
}

func (c *RouterCtx) SetSession(sess *Session) Context {
	panic("implement me")
}

func (c *RouterCtx) Request() *message.Message {
	panic("implement me")
}

func (c *RouterCtx) SetRequest(id, data interface{}) error {
	panic("implement me")
}

func (c *RouterCtx) MustSetRequest(id, data interface{}) Context {
	panic("implement me")
}

func (c *RouterCtx) SetRequestMessage(Msg *message.Message) Context {
	panic("implement me")
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

func (c *RouterCtx) MarshalResp(id uint32, flag uint16, data interface{}) error {
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
	return nil
}

func (c *RouterCtx) MustSetResponse(id, data interface{}) Context {
	panic("implement me")
}

func (c *RouterCtx) SetResponseMessage(Msg *message.Message) Context {
	panic("implement me")
}

func (c *RouterCtx) Send(id uint32, flag uint16, data interface{}) error {
	if err := c.MarshalResp(id, flag, data); err != nil {
		return err
	}

	return c.session.Send(c)
}

func (c *RouterCtx) SendTo(session *Session) {
	panic("implement me")
}

func (c *RouterCtx) Get(key string) (value interface{}, exists bool) {
	panic("implement me")
}

func (c *RouterCtx) Set(key string, value interface{}) {
	panic("implement me")
}

func (c *RouterCtx) Remove(key string) {
	panic("implement me")
}

func (c *RouterCtx) Copy() Context {
	panic("implement me")
}

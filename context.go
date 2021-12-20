package ktcp

import (
	"context"
	"fmt"
	"sync"
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
	GetRouter() *Router
	GetReqMsg() *message.Message
	Bind(v interface{}) error
	Response() *message.Message
	Send(id uint32, flag uint16, resp interface{}) error
	Middleware(middleware.Handler) middleware.Handler
	Reset(sess *Session, reqMsg *message.Message)
	Get(key string) (value interface{}, exists bool)
	Set(key string, value interface{})
}

type routerCtx struct {
	mu      sync.RWMutex
	router  *Router
	storage map[string]interface{}
	session *Session
	reqMsg  *message.Message
	respMsg *message.Message
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

// NewRouterContext returns a new Context for the given request and response.
func NewRouterContext(r *Router) *routerCtx {
	return &routerCtx{
		router: r,
	}
}

func (c *routerCtx) GetReqMsg() *message.Message {
	return c.reqMsg
}

func (c *routerCtx) GetRespMsg() *message.Message {
	return c.respMsg
}

func (c *routerCtx) GetRouter() *Router {
	return c.router
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
	return middleware.Chain(c.router.srv.ms...)(h)
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

// Get implements Context.Get method.
func (c *routerCtx) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.storage[key]
	c.mu.RUnlock()
	return
}

// Set implements Context.Set method.
func (c *routerCtx) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.storage == nil {
		c.storage = make(map[string]interface{})
	}
	c.storage[key] = value
	c.mu.Unlock()
}

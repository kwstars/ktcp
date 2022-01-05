package ktcp

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/kwstars/ktcp/message"
	"github.com/kwstars/ktcp/packing"
	"github.com/kwstars/ktcp/storage"
	"github.com/kwstars/ktcp/sync/errgroup"
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
	Send(id uint32, resp interface{}) error
	SendError(id uint32, resp interface{}) error
	Middleware(middleware.Handler) middleware.Handler
	Reset(sess *Session, reqMsg *message.Message)
	AppendToStorage(saver storage.Saver)
	Save() (err error)
}

type routerCtx struct {
	session *Session
	storage []storage.Saver
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

func (c *routerCtx) Save() (err error) {
	g := errgroup.Group{}
	for _, saver := range c.storage {
		s := saver
		g.Go(func(ctx context.Context) error {
			return s.Save(c)
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}
	return
}

func (c *routerCtx) GetReqMsg() *message.Message {
	return c.reqMsg
}

func (c *routerCtx) ForwardHandler(callback CallBack) {
	c.session.callback = callback
}

func (c *routerCtx) Reset(sess *Session, reqMsg *message.Message) {
	c.session = sess
	c.storage = c.storage[:0]
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

func (c *routerCtx) Send(id uint32, data interface{}) error {
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
		Flag: packing.OKType,
		Data: dataRaw,
	}

	return c.session.Send(c)
}

func (c *routerCtx) SendError(id uint32, data interface{}) error {

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
		Flag: packing.ErrType,
		Data: dataRaw,
	}

	return c.session.Send(c)
}

func (c *routerCtx) AppendToStorage(saver storage.Saver) {
	c.storage = append(c.storage, saver)
}

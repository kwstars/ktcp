package ktcp

import (
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

// MiddlewareFunc is the function type for middlewares.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(ctx Context) error

type HandlersChain []HandlerFunc

// Router is an HTTP router.
type Router struct {
	mux  map[uint32]HandlerFunc
	pool sync.Pool
	srv  *Server
	log  *log.Helper
}

func newRouter(log *log.Helper, srv *Server) *Router {
	r := &Router{
		mux: make(map[uint32]HandlerFunc),
		srv: srv,
		log: log,
	}

	r.pool.New = func() interface{} {
		return NewRouterContext(r)
	}

	return r
}

// GET registers a new GET route for a path with matching handler in the router.
func (r *Router) register(id uint32, h HandlerFunc, m ...HandlerFunc) {
	r.mux[id] = h
}

//func (r *Router) wrapHandlers(handler HandlerFunc) (wrapped HandlerFunc) {
//if handler == nil {
//	handler = r.notFoundHandler
//}
//if handler == nil {
//	handler = nilHandler
//}
//wrapped = handler
//for i := len(middles) - 1; i >= 0; i-- {
//	m := middles[i]
//	wrapped = m(wrapped)
//}

//	return wrapped
//}

// Handle registers a new route with a matcher for the URL path and method.
//func (r *Router) Handle(ctx Context) {
//reqMsg := ctx.GetReqMsg()

//if reqMsg == nil {
//	return
//}

//var h HandlerFunc
//if v, exist := r.mux[reqMsg.ID]; exist {
//	h = v
//} else {
//	r.log.Errorf("router not found id: %d, idType: %T", reqMsg.ID, reqMsg.ID)
//	return
//}

//TODO 中间件 recover  log tracing
//if err := h(ctx); err != nil {
//	//r.srv.ene(ctx, err)
//	r.log.Errorf("handle err: %v", err)
//}

////if v, has := r.middlewaresMapper[ctx.reqEntry.ID]; has {
////	mws = append(mws, v...) // append to global ones
////}
//
//// create the handlers stack
//wrapped := r.wrapHandlers(handler)
//
//// and call the handlers stack
//wrapped(ctx
//}

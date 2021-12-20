package main

import (
	"context"
	"fmt"

	"github.com/kwstars/ktcp/example/pb"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/kwstars/ktcp"
	"github.com/kwstars/ktcp/sync/atomic"
)

type UserService struct{}

func (s *UserService) CreateRole(ctx context.Context, req *pb.CreateRoleReq) (*pb.CreateRoleResp, error) {
	panic("implement me")
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	return &pb.LoginResp{
		Sid: 111111,
	}, nil

	//return nil, pb.ErrorUserNotFound("not found user %v", "123123123")
}

func (s *UserService) OnClose(c *ktcp.Session) {
}

func (s *UserService) OnMessage(ctx ktcp.Context) {
	fmt.Println("on userService message", ctx.GetSession().ID())

	ctx.GetRouter().Handle(ctx)
}

type Gate struct {
	Count  atomic.Int64
	Server *ktcp.Server
}

func (s *Gate) OnConnect(c *ktcp.Session) {
	s.Count.Add(1)
	fmt.Println("OnConnect:", s.Count.Get(), c.ID())
}

func (s *Gate) OnClose(c *ktcp.Session) {
	s.Count.Add(-1)
	fmt.Println("OnClose:", s.Count.Get())
}

func (s *Gate) OnMessage(ctx ktcp.Context) {
	fmt.Println("on gate message")

	// TODO 没有登陆成功需要返回错误

	// TODO 登陆成功 forward
	user := &UserService{}
	pb.RegisterUserServiceKTCPServer(s.Server, user)
	ctx.ForwardHandler(user)
}

func main() {
	// create a new server
	gate := &Gate{}
	opts := []ktcp.ServerOption{
		ktcp.Middleware(
			recovery.Recovery(),
		),
	}
	s := ktcp.NewServer(gate, opts...)
	gate.Server = s

	// start the server
	_ = s.Serve()
}

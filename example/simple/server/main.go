package main

import (
	"context"
	"fmt"

	"github.com/kwstars/ktcp/internal/sync/atomic"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/kwstars/ktcp"
	"github.com/kwstars/ktcp/example/pb"
)

type UserService struct {
	uid uint32
}

//func (s *UserService) CreateRole(ctx context.Context, req *pb.CreateRoleResponse) (*pb.CreateRoleResponse, error) {
//	panic("implement me")
//}
//
//func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
//	return &pb.LoginResponse{
//		Sid: 111111,
//	}, nil
//
//	//return nil, pb.ErrorUserNotFound("not found user %v", "123123123")
//}

func (s *UserService) CreateRole(ctx context.Context, request *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	panic("implement me")
}

func (s *UserService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{
		Sid: 111111,
	}, nil

	//return nil, pb.ErrorUserNotFound("not found user %v", "123123123")
}

func (s *UserService) OnClose(c *ktcp.Session) {
}

func (s *UserService) OnMessage(ctx ktcp.Context) {
	fmt.Printf("on userService message: %s, uid: %v\n", ctx.GetSession().ID(), s.uid)

	if err := pb.Router(ctx, s); err != nil {
		fmt.Printf("on userService message error: %v\n", err)
	}
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
	if s.Count.Get() == 1 {
		user.uid = 1
	} else if s.Count.Get() == 2 {
		user.uid = 2
	} else {
		user.uid = 3
	}

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

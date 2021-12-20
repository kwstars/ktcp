package main

import (
	"net"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/kwstars/ktcp/encoding/proto"
	v1 "github.com/kwstars/ktcp/example/pb"
	"github.com/kwstars/ktcp/message"
	"github.com/kwstars/ktcp/packing"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	packer := packing.NewDefaultPacker()
	codec := proto.New()

	go func() {
		var count int32
		for {
			var id = v1.ID_ID_LOGIN_REQUEST
			count++
			req := &v1.LoginRequest{
				Token: "aaaaa",
			}
			data, err := codec.Marshal(req)
			if err != nil {
				panic(err)
			}
			msg := &message.Message{ID: uint32(id), Data: data}
			packedMsg, err := packer.Pack(msg)
			if err != nil {
				panic(err)
			}
			if _, err := conn.Write(packedMsg); err != nil {
				panic(err)
			}
			log.Debugf("send | id: %d; size: %d; data: %s, flag: %v", id, len(data), req.String(), msg.Flag)
			time.Sleep(time.Second)
		}
	}()

	for {
		msg, err := packer.Unpack(conn)
		if err != nil {
			panic(err)
		}
		if msg.Flag == 1 {
			var respData v1.LoginResponse
			if err := codec.Unmarshal(msg.Data, &respData); err != nil {
				panic(err)
			}
			log.Infof("recv | id: %d; size: %d; data: %s, respflag: %v", msg.ID, len(msg.Data), respData.String(), msg.Flag)
		} else {
			var respData errors.Error
			if err := codec.Unmarshal(msg.Data, &respData); err != nil {
				panic(err)
			}
			log.Infof("recv | id: %d; size: %d; data: %s, respflag: %v", msg.ID, len(msg.Data), respData.String(), msg.Flag)
		}

	}
}

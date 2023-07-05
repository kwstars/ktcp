module github.com/kwstars/ktcp/cmd/protoc-gen-go-ktcp

go 1.17

replace github.com/kwstars/ktcp v0.0.1 => ../../

require (
	github.com/go-kratos/kratos/v2 v2.6.2
	github.com/kwstars/ktcp v0.0.1
	github.com/pinzolo/casee v0.0.0-20210416022938-38877fea886d
	google.golang.org/protobuf v1.29.0
)

require (
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
)

module github.com/kwstars/ktcp/cmd/protoc-gen-go-ktcp

go 1.17

replace github.com/kwstars/ktcp v0.0.1 => ../../

require (
	github.com/go-kratos/kratos/v2 v2.1.2
	github.com/kwstars/ktcp v0.0.1
	github.com/pinzolo/casee v0.0.0-20210416022938-38877fea886d
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	google.golang.org/genproto v0.0.0-20210805201207-89edb61ffb67 // indirect
	google.golang.org/grpc v1.42.0 // indirect
)

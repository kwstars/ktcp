#!/usr/bin/env bash
go build -o $GOPATH/bin/ && protoc --go-ktcp_out=. --go-ktcp_opt=paths=source_relative --go_out=. --go_opt=paths=source_relative  example/example.proto

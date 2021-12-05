package main

// notice: check your gogo protobuf version path

//go:generate protoc -I=.\proto -I=$GOPATH\pkg -I=$GOPATH\pkg\mod\github.com\gogo\protobuf@v1.3.2 --gogofaster_out=plugins=grpc:.\proto .\proto\rpc_service.proto

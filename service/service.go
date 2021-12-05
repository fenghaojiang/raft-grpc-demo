package service

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"raft-grpc-demo/core"
	rpc_servicepb "raft-grpc-demo/proto"
)

type StoreApi interface {
	Get(key string, level core.ConsistencyLevel) (string, error)

	Set(key, value string) error

	Delete(key string) error

	Join(nodeID, httpAddr, raftAddr string) error

	LeaderAPIAddr() string
}

type Service struct {
	addr    string
	store   StoreApi
	grpcSrv *grpc.Server

	logger *log.Logger
}

func New(store StoreApi, addr string) *Service {
	grpcSrv := NewGrpcServer(addr)
	return &Service{
		addr:    addr,
		store:   store,
		grpcSrv: grpcSrv,
		logger:  log.New(os.Stderr, "[grpc Service]", log.LstdFlags),
	}
}

func NewGrpcServer(addr string) *grpc.Server {
	srv := grpc.NewServer()
	rpc_servicepb.RegisterRpcServiceServer(srv, &Server{})
	network := "tcp"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("listen to network %s, address %s failed", network, addr)
	}
	go func() {
		if err := srv.Serve(ln); err != nil {
			log.Panic("socket listener accept net conn failed", err.Error())
		}
	}()
	return srv
}

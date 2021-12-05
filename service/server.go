package service

import (
	"context"
	"errors"
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

func NewServer(store StoreApi, addr string) *Server {
	return &Server{
		addr:   addr,
		store:  store,
		logger: log.New(os.Stderr, "[grpc Service]", log.LstdFlags),
	}
}

type Server struct {
	addr  string
	store StoreApi

	logger *log.Logger
}

var _ rpc_servicepb.RpcServiceServer = (*Server)(nil) // 检查是否实现所有方法

func NewGrpcServer(addr string, api StoreApi) *grpc.Server {
	grpcSrv := grpc.NewServer()
	srv := NewServer(api, addr)
	rpc_servicepb.RegisterRpcServiceServer(grpcSrv, srv)
	network := "tcp"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("listen to network %s, address %s failed", network, addr)
	}
	go func() {
		if err := grpcSrv.Serve(ln); err != nil {
			log.Panic("socket listener accept net conn failed", err.Error())
		}
	}()
	return grpcSrv
}

func (s *Server) Get(ctx context.Context, req *rpc_servicepb.GetReq) (*rpc_servicepb.GetRsp, error) {
	if req.Key == "" {
		return nil, errors.New("")
	}
	return nil, nil
}

func (s *Server) Set(ctx context.Context, req *rpc_servicepb.SetReq) (*rpc_servicepb.SetRsp, error) {
	//TODO
	return nil, nil
}

func (s *Server) Delete(ctx context.Context, req *rpc_servicepb.DeleteReq) (*rpc_servicepb.DeleteRsp, error) {
	//TODO
	return nil, nil
}

func (s *Server) Join(ctx context.Context, req *rpc_servicepb.JoinReq) (*rpc_servicepb.JoinRsp, error) {
	//TODO
	return nil, nil
}

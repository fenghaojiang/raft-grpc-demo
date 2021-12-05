package service

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	rpc_servicepb "raft-grpc-demo/proto"
)

type Server struct {
	service *Service
}

var _ rpc_servicepb.RpcServiceServer = (*Server)(nil) // 检查是否实现所有方法

func NewGrpcServer(addr string, api StoreApi) *grpc.Server {
	srv := grpc.NewServer()
	service := NewService(api, addr)
	rpc_servicepb.RegisterRpcServiceServer(srv, &Server{service: service})
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

func (s *Server) Get(ctx context.Context, req *rpc_servicepb.GetReq) (*rpc_servicepb.GetRsp, error) {
	//TODO
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

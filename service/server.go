package service

import (
	"context"
	rpc_servicepb "raft-grpc-demo/proto"
)

type Server struct {
}

var _ rpc_servicepb.RpcServiceServer = (*Server)(nil) // 检查是否实现所有方法

func (s *Server) Get(ctx context.Context, req *rpc_servicepb.GetReq) (*rpc_servicepb.GetRsp, error) {
	return nil, nil
}

func (s *Server) Set(ctx context.Context, req *rpc_servicepb.SetReq) (*rpc_servicepb.SetRsp, error) {
	return nil, nil
}

func (s *Server) Delete(ctx context.Context, req *rpc_servicepb.DeleteReq) (*rpc_servicepb.DeleteRsp, error) {
	return nil, nil
}

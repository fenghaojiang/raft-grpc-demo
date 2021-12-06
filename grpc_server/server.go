package grpc_server

import (
	"context"
	"log"
	"net"
	"os"
	"raft-grpc-demo/core"
	"raft-grpc-demo/error_code"
	rpcservicepb "raft-grpc-demo/proto"

	"google.golang.org/grpc"
)

type StoreApi interface {
	Get(key string, level core.ConsistencyLevel) (string, error)

	Set(key, value string) error

	Delete(key string) error

	Join(nodeID, httpAddr, raftAddr string) error

	LeaderAPIAddr() string
}

func NewServer(store StoreApi, addr string, ln net.Listener) *Server {
	return &Server{
		addr:   addr,
		store:  store,
		ln:     ln,
		logger: log.New(os.Stderr, "[grpc Service]", log.LstdFlags),
	}
}

type Server struct {
	addr   string
	store  StoreApi
	ln     net.Listener
	logger *log.Logger
	//leadClient rpcservicepb.RpcServiceClient
	leaderConn *grpc.ClientConn
}

var _ rpcservicepb.RpcServiceServer = (*Server)(nil) // 检查是否实现所有方法
var rpcserviceClient rpcservicepb.RpcServiceClient

func NewGrpcServer(addr string, api StoreApi) *grpc.Server {
	grpcSrv := grpc.NewServer()
	ln, err := net.Listen("tcp", addr)
	srv := NewServer(api, addr, ln)
	rpcservicepb.RegisterRpcServiceServer(grpcSrv, srv)
	network := "tcp"
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

func (s *Server) Close() {
	s.ln.Close()
}

func (s *Server) Get(ctx context.Context, req *rpcservicepb.GetReq) (*rpcservicepb.GetRsp, error) {
	if req.Key == "" {
		return nil, error_code.BadRequest
	}
	var consLv core.ConsistencyLevel
	switch req.Level {
	case "default":
		consLv = core.Default
	case "stale":
		consLv = core.Stale
	case "consistent":
		consLv = core.Consistent
	default:
		consLv = core.Default
	}
	value, err := s.store.Get(req.Key, consLv)
	if err != nil {
		if err == core.ErrNotLeader {

		}
		return nil, error_code.InternalServerError
	}
	return &rpcservicepb.GetRsp{Value: value}, nil
}

func (s *Server) get(ctx context.Context, leaderGrpcAddr, key string) (*rpcservicepb.GetRsp, error) {
	var err error
	s.leaderConn, err = grpc.Dial(leaderGrpcAddr)
	if err != nil {
		return nil, err
	}
	rpcserviceClient = rpcservicepb.NewRpcServiceClient(s.leaderConn)
	rsp, err := rpcserviceClient.Get(ctx, &rpcservicepb.GetReq{Key: key})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

//TODO
// func (s *Server) verifyLeaderConnAndDosomething(ctx context.Context, req interface{}, f func(args ...string) interface{}, error) (interface{}, error) {
// 	leaderGrpcAddr := s.store.LeaderAPIAddr()
// 	if leaderGrpcAddr == "" {
// 		return nil, error_code.ServiceUnavailable
// 	}
// 	if s.leaderConn == nil {
// 		rsp, err := f
// 		if err != nil {
// 			return nil, err
// 		}
// 		return rsp, nil
// 	} else {
// 		if leaderGrpcAddr == s.leaderConn.Target() {
// 			rpcserviceClient.Get(ctx, &rpcservicepb.GetReq{Key: req.Key})
// 		} else {
// 			rsp, err := s.get(ctx, leaderGrpcAddr, req.Key)
// 			if err != nil {
// 				return nil, err
// 			}
// 			return rsp, nil
// 		}
// 	}
// }

func (s *Server) Set(ctx context.Context, req *rpcservicepb.SetReq) (*rpcservicepb.SetRsp, error) {
	if err := s.store.Set(req.Key, req.Value); err != nil {
		if err == core.ErrNotLeader {

		}
		return nil, err
	}
	return &rpcservicepb.SetRsp{}, nil
}

func (s *Server) Delete(ctx context.Context, req *rpcservicepb.DeleteReq) (*rpcservicepb.DeleteRsp, error) {
	//TODO
	return nil, nil
}

func (s *Server) Join(ctx context.Context, req *rpcservicepb.JoinReq) (*rpcservicepb.JoinRsp, error) {
	//TODO
	return nil, nil
}

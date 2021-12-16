package grpc_server

import (
	"context"
	"log"
	"net"
	"os"
	"raft-grpc-demo/core"
	"raft-grpc-demo/error_code"
	rpcservicepb "raft-grpc-demo/proto"
	"time"

	"google.golang.org/grpc"
)

type StoreApi interface {
	Get(key string, level core.ConsistencyLevel) (string, error)

	Set(key, value string) error

	Delete(key string) error

	Join(nodeID, grpcAddr, raftAddr string) error

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

const (
	GetTypeID = int64(iota)
	SetTypeID
	JoinTypeID
	DeleteTypeID
)

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

func NewGrpcServerAndStart(addr string, api StoreApi) error {
	grpcSrv := grpc.NewServer()
	network := "tcp"
	ln, err := net.Listen(network, addr)
	srv := NewServer(api, addr, ln)
	rpcservicepb.RegisterRpcServiceServer(grpcSrv, srv)
	if err != nil {
		return err
	}
	go func() {
		if err := grpcSrv.Serve(ln); err != nil {
			log.Panic("socket listener accept net conn failed", err.Error())
		}
	}()
	return nil
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
			rsp, err := s.verifyLeaderConnReDial(ctx, req, GetTypeID)
			if err != nil {
				return nil, err
			}
			return rsp.(*rpcservicepb.GetRsp), nil
		}
		return nil, error_code.InternalServerError
	}
	return &rpcservicepb.GetRsp{Value: value}, nil
}

func (s *Server) get(ctx context.Context, leaderGrpcAddr, key string) (*rpcservicepb.GetRsp, error) {
	var err error
	timeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	s.leaderConn, err = grpc.DialContext(timeCtx, leaderGrpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	rpcserviceClient = rpcservicepb.NewRpcServiceClient(s.leaderConn)
	rsp, err := rpcserviceClient.Get(timeCtx, &rpcservicepb.GetReq{Key: key})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *Server) verifyLeaderConnReDial(ctx context.Context, req interface{}, typeID int64) (interface{}, error) {
	leaderGrpcAddr := s.store.LeaderAPIAddr()
	if leaderGrpcAddr == "" {
		return nil, error_code.ServiceUnavailable
	}
	switch typeID {
	case GetTypeID:
		if s.leaderConn == nil {
			rsp, err := s.get(ctx, leaderGrpcAddr, req.(*rpcservicepb.GetReq).Key)
			if err != nil {
				return nil, err
			}
			return rsp, err
		} else {
			if leaderGrpcAddr == s.leaderConn.Target() {
				rsp, err := rpcserviceClient.Get(ctx, req.(*rpcservicepb.GetReq))
				if err != nil {
					return nil, err
				}
				return rsp, nil
			} else {
				rsp, err := s.get(ctx, leaderGrpcAddr, req.(*rpcservicepb.GetReq).Key)
				if err != nil {
					return nil, err
				}
				return rsp, nil
			}
		}
	case SetTypeID:
		if s.leaderConn == nil {
			rsp, err := s.set(ctx, leaderGrpcAddr, req.(*rpcservicepb.SetReq).Key, req.(*rpcservicepb.SetReq).Value)
			if err != nil {
				return nil, err
			}
			return rsp, err
		} else {
			if leaderGrpcAddr == s.leaderConn.Target() {
				rsp, err := rpcserviceClient.Set(ctx, req.(*rpcservicepb.SetReq))
				if err != nil {
					return nil, err
				}
				return rsp, nil
			} else {
				rsp, err := s.set(ctx, leaderGrpcAddr, req.(*rpcservicepb.SetReq).Key, req.(*rpcservicepb.SetReq).Value)
				if err != nil {
					return nil, err
				}
				return rsp, nil
			}
		}
	case DeleteTypeID:
		if s.leaderConn == nil {
			rsp, err := s.delete(ctx, leaderGrpcAddr, req.(*rpcservicepb.DeleteReq).Key)
			if err != nil {
				return nil, err
			}
			return rsp, err
		} else {
			if leaderGrpcAddr == s.leaderConn.Target() {
				rsp, err := rpcserviceClient.Delete(ctx, req.(*rpcservicepb.DeleteReq))
				if err != nil {
					return nil, err
				}
				return rsp, nil
			} else {
				rsp, err := s.delete(ctx, leaderGrpcAddr, req.(*rpcservicepb.DeleteReq).Key)
				if err != nil {
					return nil, err
				}
				return rsp, nil
			}
		}
	case JoinTypeID:
		if s.leaderConn == nil {
			rsp, err := s.join(ctx, leaderGrpcAddr, req.(*rpcservicepb.JoinReq).GrpcAddr, req.(*rpcservicepb.JoinReq).RaftAddr, req.(*rpcservicepb.JoinReq).NodeID)
			if err != nil {
				return nil, err
			}
			return rsp, err
		} else {
			if leaderGrpcAddr == s.leaderConn.Target() {
				rsp, err := rpcserviceClient.Join(ctx, req.(*rpcservicepb.JoinReq))
				if err != nil {
					return nil, err
				}
				return rsp, nil
			} else {
				rsp, err := s.join(ctx, leaderGrpcAddr, req.(*rpcservicepb.JoinReq).GrpcAddr, req.(*rpcservicepb.JoinReq).RaftAddr, req.(*rpcservicepb.JoinReq).NodeID)
				if err != nil {
					return nil, err
				}
				return rsp, nil
			}
		}
	default:
		return nil, error_code.NoTypeIDError
	}

}

func (s *Server) Set(ctx context.Context, req *rpcservicepb.SetReq) (*rpcservicepb.SetRsp, error) {
	if err := s.store.Set(req.Key, req.Value); err != nil {
		if err == core.ErrNotLeader {
			rsp, err := s.verifyLeaderConnReDial(ctx, req, SetTypeID)
			if err != nil {
				return nil, err
			}
			return rsp.(*rpcservicepb.SetRsp), nil
		}
		return nil, err
	}
	return &rpcservicepb.SetRsp{}, nil
}

func (s *Server) set(ctx context.Context, leaderGrpcAddr, key, value string) (interface{}, error) {
	var err error
	timeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	s.leaderConn, err = grpc.DialContext(timeCtx, leaderGrpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	rpcserviceClient = rpcservicepb.NewRpcServiceClient(s.leaderConn)
	rsp, err := rpcserviceClient.Set(timeCtx, &rpcservicepb.SetReq{Key: key, Value: value})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *Server) Delete(ctx context.Context, req *rpcservicepb.DeleteReq) (*rpcservicepb.DeleteRsp, error) {
	if err := s.store.Delete(req.Key); err != nil {
		if err == core.ErrNotLeader {
			rsp, err := s.verifyLeaderConnReDial(ctx, req, DeleteTypeID)
			if err != nil {
				return nil, err
			}
			return rsp.(*rpcservicepb.DeleteRsp), nil
		}
		return nil, err
	}
	return &rpcservicepb.DeleteRsp{}, nil
}

func (s *Server) delete(ctx context.Context, leaderGrpcAddr, key string) (interface{}, error) {
	var err error
	timeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	s.leaderConn, err = grpc.DialContext(timeCtx, leaderGrpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	rpcserviceClient = rpcservicepb.NewRpcServiceClient(s.leaderConn)
	rsp, err := rpcserviceClient.Delete(timeCtx, &rpcservicepb.DeleteReq{Key: key})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *Server) Join(ctx context.Context, req *rpcservicepb.JoinReq) (*rpcservicepb.JoinRsp, error) {
	if err := s.store.Join(req.NodeID, req.GrpcAddr, req.RaftAddr); err != nil {
		if err == core.ErrNotLeader {
			rsp, err := s.verifyLeaderConnReDial(ctx, req, JoinTypeID)
			if err != nil {
				return nil, err
			}
			return rsp.(*rpcservicepb.JoinRsp), nil
		}
		return nil, err
	}
	return &rpcservicepb.JoinRsp{}, nil
}

func (s *Server) join(ctx context.Context, leaderGrpcAddr, grpcAddr, raftAddr, nodeID string) (interface{}, error) {
	var err error
	timeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	s.leaderConn, err = grpc.DialContext(timeCtx, leaderGrpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	rpcserviceClient = rpcservicepb.NewRpcServiceClient(s.leaderConn)
	rsp, err := rpcserviceClient.Join(timeCtx, &rpcservicepb.JoinReq{GrpcAddr: grpcAddr, RaftAddr: raftAddr, NodeID: nodeID})
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"raft-grpc-demo/core"
	"raft-grpc-demo/grpc_server"
	rpcservicepb "raft-grpc-demo/proto"
	"time"
)

var (
	grpcAddr     = flag.String("svc", "localhost:51000", "service host:port for this node")
	raftId       = flag.String("id", "", "node id used by Raft")
	raftDataDir  = flag.String("data", "data/", "raft data dir")
	raftAddr     = flag.String("raft", "localhost:52000", "raft host:port for this node")
	joinAddr     = flag.String("join", "", "join address")
	registerAddr = flag.String("service_join", "localhost:50000", "raft register center port")
)

func main() {
	flag.Parse()

	if *raftId == "" {
		log.Fatalf("raft id is required")
	}
	os.MkdirAll(*raftDataDir, 0700)

	s := core.NewStore()
	s.RaftAddr = *raftAddr
	s.RaftId = *raftId
	s.RaftDataDir = *raftDataDir

	if err := s.StartRaft(*joinAddr == ""); err != nil {
		log.Fatalf("s.StartRaft: %v", err)
	}
	if *joinAddr != "" {
		if err := join(*joinAddr, *grpcAddr, *raftAddr, *raftId); err != nil {
			log.Fatalf("failed to join node at %s: %s", *joinAddr, err.Error())
		}
	} else {
		log.Println("no join addresses set")
	}

	// Wait until the store is in full consensus.
	openTimeout := 120 * time.Second
	s.WaitForLeader(openTimeout)
	s.WaitForApplied(openTimeout)

	if err := s.SetMeta(*raftId, *grpcAddr); err != nil && err != core.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to SetMeta at %s: %s", *raftId, err.Error())
	}

	grpc_server.NewGrpcServerAndStart(*grpcAddr, s)

	b, err := json.Marshal(map[string]string{"serviceAddr": *grpcAddr})
	resp, err := http.Post(fmt.Sprintf("http://%s/service_join", *registerAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		log.Fatalf("join service to client fail %s", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("join service to client fail %d", resp.StatusCode)
	}

	log.Println("started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("exiting")

}

func join(joinAddr, grpcAddr, raftAddr, nodeID string) error {
	ctx := context.Background()
	cc, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer cc.Close()
	rpcClient := rpcservicepb.NewRpcServiceClient(cc)
	_, err = rpcClient.Join(ctx, &rpcservicepb.JoinReq{
		GrpcAddr: grpcAddr,
		RaftAddr: raftAddr,
		NodeID:   nodeID,
	})
	if err != nil {
		return err
	}
	return nil
}

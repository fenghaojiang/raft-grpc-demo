package main

import (
	"flag"
	"log"
	"os"
	"raft-grpc-demo/core"
)

var (
	grpcAddr     = flag.String("svc", "localhost:51000", "service host:port for this node")
	raftId       = flag.String("id", "", "node id used by Raft")
	raftDataDir  = flag.String("data", "data/", "raft data dir")
	raftAddr     = flag.String("raft", "localhost:52000", "raft host:port for this node")
	joinAddr     = flag.String("join", "", "join address")
	registerAddr = flag.String("service_join", "localhost:50000", "raft client port")
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
}

package main

import "flag"

var (
	httpAddr    = flag.String("svc", "localhost:51000", "service host:port for this node")
	raftId      = flag.String("id", "", "node id used by Raft")
	raftDataDir = flag.String("data", "data/", "raft data dir")
	raftAddr    = flag.String("raft", "localhost:52000", "raft host:port for this node")
	joinAddr    = flag.String("join", "", "join address")
	clientAddr  = flag.String("service_join", "localhost:50000", "raft client port")
)

func main() {

}

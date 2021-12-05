package service

import (
	"log"
	"os"
	"raft-grpc-demo/core"
)

type StoreApi interface {
	Get(key string, level core.ConsistencyLevel) (string, error)

	Set(key, value string) error

	Delete(key string) error

	Join(nodeID, httpAddr, raftAddr string) error

	LeaderAPIAddr() string
}

type Service struct {
	addr  string
	store StoreApi

	logger *log.Logger
}

func NewService(store StoreApi, addr string) *Service {
	return &Service{
		addr:   addr,
		store:  store,
		logger: log.New(os.Stderr, "[grpc Service]", log.LstdFlags),
	}
}

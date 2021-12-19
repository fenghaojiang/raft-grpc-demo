package main

import (
	"log"
	"os"
	"os/signal"
	"raft-grpc-demo/register"
)

func main() {
	registerCenter := register.NewCenterForRegister("127.0.0.1:50000")
	err := registerCenter.Start()
	if err != nil {
		log.Fatalf("raft register center start fail")
	}
	log.Println("raft register center start success")
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("exiting")
}

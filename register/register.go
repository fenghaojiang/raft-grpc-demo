package register

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"raft-grpc-demo/ecode"
	rpcservicepb "raft-grpc-demo/proto"
	"strings"
	"time"

	"google.golang.org/grpc"
)

type centerForRegister struct {
	addr     string
	conn     *grpc.ClientConn
	services map[string]struct{}
	ln       net.Listener
	logger   *log.Logger
}

var rpcClient rpcservicepb.RpcServiceClient

//NewCenterForRegister initialize registerCenter
func NewCenterForRegister(addr string) *centerForRegister {
	return &centerForRegister{
		addr:     addr,
		services: map[string]struct{}{},
		logger:   log.New(os.Stderr, "[RegisterCenter Service]", log.LstdFlags),
	}
}

func (c *centerForRegister) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	getKey := func() string {
		parts := strings.Split(req.URL.Path, "/")
		if len(parts) != 3 {
			return ""
		}
		return parts[2]
	}
	if strings.HasPrefix(req.URL.Path, "/key") {
		switch req.Method {
		case "GET":
			k := getKey()
			if k == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			v, err := c.doGet(k)
			if err != nil {
				c.logger.Printf("get key %s fail %v", k, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			b, err := json.Marshal(map[string]string{k: v})
			if err != nil {
				c.logger.Printf("marshal key %s fail %v", k, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			io.WriteString(w, string(b))
		case "POST":
			m := map[string]string{}
			if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			for key := range m {
				err := c.doSet(key, m[key])
				if err != nil {
					c.logger.Printf("set key %s fail %v", key, err)
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			w.WriteHeader(http.StatusOK)
		case "DELETE":
			k := getKey()
			if k == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err := c.doDelete(k)
			if err != nil {
				c.logger.Printf("delete key %s fail %v", k, err)
				io.WriteString(w, err.Error())
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(req.URL.Path, "/service_join") {
		c.serviceRegister(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *centerForRegister) dialRegisteredAddress() error {
	var err error
	var targetAddr = ""
	if len(c.services) == 1 {
		for serviceAddr := range c.services {
			targetAddr += serviceAddr
		}
		timeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		c.conn, err = grpc.DialContext(timeCtx, targetAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"pick_first"}`))
		if err != nil {
			return err
		}
		if c.conn == nil {
			fmt.Println("dial ", targetAddr, "fail")
			return ecode.ErrNoAvailableService
		}
		return nil
	}
	targetAddr = "static:///"
	var cnt int
	for serviceAddr := range c.services {
		targetAddr += serviceAddr
		cnt++
		if cnt != len(c.services) {
			targetAddr += ","
		}
	}
	fmt.Println(targetAddr)
	timeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c.conn, err = grpc.DialContext(timeCtx, targetAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"pick_first"}`))
	if err != nil {
		return err
	}
	if c.conn == nil {
		fmt.Println("dial ", targetAddr, "fail")
		return ecode.ErrNoAvailableService
	}
	return nil
}

func (c *centerForRegister) serviceRegister(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		m := map[string]string{}
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		addr := m["serviceAddr"]
		c.addService(addr)
		err := c.dialRegisteredAddress()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		rpcClient = rpcservicepb.NewRpcServiceClient(c.conn)
		log.Printf("server join addr: %s", addr)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

}

func (c *centerForRegister) Start() error {
	if len(c.addr) == 0 {
		log.Fatalf("raft client addr is required")
	}
	server := http.Server{
		Handler: c,
	}
	ln, err := net.Listen("tcp", c.addr)
	if err != nil {
		log.Printf("init listener fail")
		return err
	}
	c.ln = ln

	http.Handle("/", c)
	go func() {
		err := server.Serve(c.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()
	return nil
}

func (c *centerForRegister) addService(addr string) {
	c.services[addr] = struct{}{}
}

func (c *centerForRegister) removeService(addr string) {
	delete(c.services, addr)
}

func (c *centerForRegister) doGet(key string) (string, error) {
	if c.conn == nil {
		err := c.dialRegisteredAddress()
		if err != nil {
			return "", err
		}
	}
	if rpcClient == nil {
		rpcClient = rpcservicepb.NewRpcServiceClient(c.conn)
	}
	rsp, err := rpcClient.Get(context.Background(), &rpcservicepb.GetReq{Key: key})
	if err != nil {
		return "", err
	}
	return rsp.Value, err
}

func (c *centerForRegister) doSet(key string, value string) error {
	if rpcClient == nil {
		rpcClient = rpcservicepb.NewRpcServiceClient(c.conn)
	}
	_, err := rpcClient.Set(context.Background(), &rpcservicepb.SetReq{Key: key, Value: value})
	if err != nil {
		return err
	}
	return nil
}

func (c *centerForRegister) doDelete(key string) error {
	if rpcClient == nil {
		rpcClient = rpcservicepb.NewRpcServiceClient(c.conn)
	}
	_, err := rpcClient.Delete(context.Background(), &rpcservicepb.DeleteReq{Key: key})
	if err != nil {
		return err
	}
	return nil
}

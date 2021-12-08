package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	rpcservicepb "raft-grpc-demo/proto"
	"strings"

	"google.golang.org/grpc"
)

const maxFailOnRequestOnService int64 = 5

type CenterForRegister struct {
	addr     string
	conn     *grpc.ClientConn
	services map[string]struct{}
	ln       net.Listener
	logger   *log.Logger
}

var rpcClient rpcservicepb.RpcServiceClient

func NewCenterForRegister(addr string) *CenterForRegister {
	return &CenterForRegister{
		addr:     addr,
		services: map[string]struct{}{},
	}
}

func (c *CenterForRegister) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			b, err := json.Marshal(map[string]string{k: v})
			if err != nil {
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
			err := c.doPost(m)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case "DELETE":
			k := getKey()
			if k == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			c.doDelete(k)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if strings.HasPrefix(req.URL.Path, "/service_join") {
		c.serviceRegister(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (c *CenterForRegister) serviceRegister(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		m := map[string]string{}
		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if addr, ok := m["serviceAddr"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			c.services[addr] = struct{}{}
			var err error
			var targetAddr = "static://"
			var cnt int
			for serviceAddr := range c.services {
				targetAddr += serviceAddr
				cnt++
				if cnt != len(c.services) {
					targetAddr += ","
				}
			}
			c.conn, err = grpc.DialContext(context.Background(), targetAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultServiceConfig("round_robin"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			rpcClient = rpcservicepb.NewRpcServiceClient(c.conn)
			log.Printf("server join addr: %s", addr)
			w.WriteHeader(http.StatusOK)
			return
		}
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

}

func (c *CenterForRegister) Start() error {
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

func (c *CenterForRegister) addService(addr string) {
	c.services[addr] = struct{}{}
}

func (c *CenterForRegister) removeService(addr string) {
	delete(c.services, addr)
}

func (c *CenterForRegister) doGet(key string) (string, error) {
	rsp, err := rpcClient.Get(context.Background(), &rpcservicepb.GetReq{Key: key})
	if err != nil {
		return "", err
	}

	return rsp.Value, err
}

func (c *CenterForRegister) doPost(m map[string]string) error {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("failed to encode key and value for POST: %s", err)
		return err
	}
	for serviceAddr := range c.services {
		resp, err := c.h.Post(fmt.Sprintf("http://%s/key", serviceAddr), "application-type/json", bytes.NewReader(b))
		if err != nil {
			log.Printf("failed to encode key and value for POST: %s", err)
			if c.services[serviceAddr] > maxFailOnRequestOnService {
				c.removeService(serviceAddr)
			} else {
				c.services[serviceAddr]++
			}
			continue
		}
		resp.Body.Close()
		return nil
	}
	return ErrNoAvailableService

}

func (c *CenterForRegister) doSet(key string, value string) error {

	return nil
}

func (c *CenterForRegister) doDelete(key string) {
	for serviceAddr := range c.services {
		ru, err := url.Parse(fmt.Sprintf("http://%s/key/%s", serviceAddr, key))
		if err != nil {
			log.Printf("failed to parse URL for delete: %s", err)
			continue
		}
		req := &http.Request{
			Method: "DELETE",
			URL:    ru,
		}
		resp, err := c.h.Do(req)
		if err != nil {
			log.Printf("failed to GET key: %s", err)
			if c.services[serviceAddr] > maxFailOnRequestOnService {
				c.removeService(serviceAddr)
			} else {
				c.services[serviceAddr]++
			}
		}
		defer resp.Body.Close()
	}
}

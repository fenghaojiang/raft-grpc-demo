package register

import (
	"fmt"
	"google.golang.org/grpc/resolver"
	"strings"
	"sync"
)

type StaticResolverBuilder struct{}

func init() {
	resolver.Register(&StaticResolverBuilder{})
}

func (srb *StaticResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 解析target.Endpoint (例如：localhost:50051,localhost:50052,localhost:50053)
	endpoints := strings.Split(target.Endpoint, ",")

	r := &StaticResolver{
		endpoints: endpoints,
		cc:        cc,
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (srb *StaticResolverBuilder) Scheme() string {
	return "static"
}

type StaticResolver struct {
	endpoints []string
	cc        resolver.ClientConn
	sync.Mutex
}

func (sr *StaticResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	sr.Lock()
	sr.resolve()
	sr.Unlock()
}

func (sr *StaticResolver) resolve() {
	var resolveAddr []resolver.Address
	for i, addr := range sr.endpoints {
		resolveAddr = append(resolveAddr, resolver.Address{
			Addr:       addr,
			ServerName: fmt.Sprintf("grpc-server-nodeID-%d", i+1),
		})
	}
	newState := resolver.State{
		Addresses: resolveAddr,
	}

	sr.cc.UpdateState(newState)
}

func (sr *StaticResolver) Close() {

}

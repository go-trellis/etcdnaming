// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

var defaultMapResolvers = &mapResolvers{Resolvers: make(map[string]*Resolver, 0)}

type mapResolvers struct {
	Resolvers map[string]*Resolver
	locker    sync.RWMutex
}

func (p *mapResolvers) getResolvers(name string) (*Resolver, bool) {
	p.locker.RLock()
	defer p.locker.RUnlock()
	r, ok := p.Resolvers[name]
	if ok {
		return r, true
	}
	return nil, false
}

func (p *mapResolvers) setResolvers(name string, r *Resolver) bool {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.Resolvers[name] = r
	return true
}

// GetResolverConn get resolver connection
func GetResolverConn(name string) (*grpc.ClientConn, bool) {
	r, ok := defaultMapResolvers.getResolvers(name)
	if !ok {
		return nil, false
	}
	if r.conn == nil {
		return nil, false
	}
	return r.conn, true
}

// Resolver is the implementaion of grpc.naming.Resolver
type Resolver struct {
	// service name to resolve
	serviceName string
	// etcd target
	etcdTarget string
	// timeout
	timeout time.Duration
	// grpc conn
	client *clientv3.Client

	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

// NewResolver return resolver with service name
// serviceName server's name
// target: etcd client server address
func NewResolver(serviceName, target string, timeout time.Duration) error {

	if _, ok := defaultMapResolvers.getResolvers(serviceName); ok {
		return nil
	}
	// generate etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		return fmt.Errorf("grpclib: creat clientv3 client failed: %s", err.Error())
	}
	_ctx, _cancel := context.WithTimeout(context.Background(), timeout)
	r := &Resolver{serviceName: serviceName, client: client, cancel: _cancel}

	conn, err := grpc.DialContext(_ctx, target, grpc.WithInsecure(), grpc.WithBalancer(grpc.RoundRobin(r)))
	if err != nil {
		return err
	}
	r.conn = conn

	if defaultMapResolvers.setResolvers(serviceName, r) {
		return nil
	}

	return fmt.Errorf("grpclib: creat clientv3 client failed: set resolver, %s", serviceName)
}

// Resolve to resolve the service from etcd, target is the dial address of etcd
// target example: "http://127.0.0.1:2379,http://127.0.0.1:12379,http://127.0.0.1:22379"
func (p *Resolver) Resolve(target string) (naming.Watcher, error) {
	if 0 == len(p.serviceName) {
		return nil, errors.New("grpclib: no service name provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	w := &watcher{
		resolver: p,
		target:   fmt.Sprintf("/%s/%s/", prefix, p.serviceName),
		ctx:      ctx,
		cancel:   cancel}
	// Return watcher
	return w, nil
}

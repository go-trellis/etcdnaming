// GNU GPL v3 License
// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/resolver"
)

// Build 生成resolver
func (p *etcdBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {

	// generate etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(p.opts.Endpoint, ","),
	})
	if err != nil {
		return nil, err
	}

	r := &etcdResolver{
		opts: p.opts,

		client: client,
		cc:     cc,
		target: target,
		rn:     make(chan struct{}, 1),

		key: fmt.Sprintf("/%s/%s/%s", p.Scheme(), p.opts.Server, p.opts.Version),
	}

	r.ctx, r.cancel = context.WithCancel(context.Background())

	go r.watcher()
	r.ResolveNow(resolver.ResolveNowOptions{})

	return r, nil
}

func (p *etcdResolver) Close() {
	p.cancel()
}

func (p *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case p.rn <- struct{}{}:
	default:
	}
}

func (p *etcdResolver) watcher() {
	resp, err := p.client.Get(p.ctx, p.key, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return
	}

	addrs := p.extractAddrs(resp)
	p.cc.UpdateState(resolver.State{
		Addresses: addrs,
	})
	rch := p.client.Watch(context.Background(), p.key, clientv3.WithPrefix(), clientv3.WithRev(resp.Header.Revision+1))
	t := time.NewTimer(p.opts.LooperTime)
	for {

		select {
		case <-t.C:
		case n := <-rch:
			for _, ev := range n.Events {
				switch ev.Type {
				case mvccpb.PUT, mvccpb.DELETE:
					resp, err := p.client.Get(p.ctx, p.key, clientv3.WithPrefix(), clientv3.WithSerializable())
					if err != nil {
						break
					}

					addrs := p.extractAddrs(resp)

					p.cc.UpdateState(resolver.State{
						Addresses: addrs,
					})
				}
			}
		case <-p.ctx.Done():
			t.Stop()
			return
		}
	}
}

func (p *etcdResolver) extractAddrs(resp *clientv3.GetResponse) []resolver.Address {
	var adds []resolver.Address
	for i := range resp.Kvs {
		adds = append(adds, resolver.Address{
			Addr: strings.TrimPrefix(string(resp.Kvs[i].Key), p.key+"/"),
		})
	}

	return adds
}

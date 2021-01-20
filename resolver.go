/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

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

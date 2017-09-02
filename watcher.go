// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import (
	"github.com/coreos/etcd/clientv3"
	etcdNaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/naming"
)

// watcher is the implementaion of grpc.naming.Watcher
type watcher struct {
	resolver *Resolver
	target   string
	ctx      context.Context
	cancel   context.CancelFunc
	wch      clientv3.WatchChan
	err      error
}

// Next to return the updates
func (w *watcher) Next() ([]*naming.Update, error) {

	if w.wch == nil {
		// first Next() returns all addresses
		return w.firstNext()
	}

	if w.err != nil {
		return nil, w.err
	}

	// process new events on target/*
	wr, ok := <-w.wch
	if !ok {
		w.err = grpc.Errorf(codes.Unavailable, "%s", etcdNaming.ErrWatcherClosed)
		return nil, w.err
	}
	if w.err = wr.Err(); w.err != nil {
		return nil, w.err
	}

	updates := make([]*naming.Update, 0, len(wr.Events))
	for _, ev := range wr.Events {
		switch ev.Type {
		case mvccpb.PUT:
			return []*naming.Update{{Op: naming.Add, Addr: string(ev.Kv.Value)}}, nil
		case mvccpb.DELETE:
			return []*naming.Update{{Op: naming.Delete, Addr: string(ev.Kv.Value)}}, nil
		}
	}
	return updates, nil
}

func extractAddrs(resp *clientv3.GetResponse) []string {
	if resp == nil || resp.Kvs == nil {
		return nil
	}

	addrs := []string{}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

func (w *watcher) firstNext() ([]*naming.Update, error) {
	// Use serialized request so resolution still works if the target etcd
	// server is partitioned away from the quorum.
	resp, err := w.resolver.client.Get(w.ctx, w.target, clientv3.WithPrefix(), clientv3.WithSerializable())
	if w.err = err; err != nil {
		return nil, err
	}

	addrs := extractAddrs(resp)
	updates := make([]*naming.Update, len(addrs))
	if l := len(addrs); l != 0 {
		for i := range addrs {
			updates[i] = &naming.Update{Op: naming.Add, Addr: addrs[i]}
		}
	}

	w.wch = w.resolver.client.Watch(w.ctx, w.target,
		[]clientv3.OpOption{
			clientv3.WithRev(resp.Header.Revision + 1),
			clientv3.WithPrefix(),
			clientv3.WithPrevKV(),
		}...)
	return updates, nil
}

func (w *watcher) Close() {
	w.resolver.cancel()
	w.cancel()
}

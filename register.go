// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"golang.org/x/net/context"
)

// Regist server regist into etcd
func (p *defaultServerRegister) Regist() (err error) {
	// get endpoints for register dial address
	if p.client, err = clientv3.New(clientv3.Config{
		Endpoints: strings.Split(p.src.Target, ","),
	}); err != nil {
		return fmt.Errorf("grpclib: create clientv3 client failed: %v", err)
	}

	go func() {
		count := 0
		for {
			if err = p.regist(); err != nil {
				count++
				if p.src.RegistRetryTimes < count {
					panic(fmt.Errorf("%s regist into etcd failed times above: %d, %v", p.src.Name, count, err))
				}
				continue
			}
			count = 0
			select {
			case <-p.stopSignal:
				return
			case <-p.ticker.C:
			}
		}
	}()

	// getting stop message and revoke etcd
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		message := <-ch
		log.Printf("receive signal '%v'", message)
		if err := p.Revoke(); err != nil {
			log.Printf("failure to revoke etcd: '%v'", err)
		} else {
			log.Printf("revoke etcd ok")
		}
	}()

	return
}

func (p *defaultServerRegister) regist() (err error) {
	// minimum lease TTL is ttl-second
	resp, ie := p.client.Grant(context.TODO(), int64(p.src.TTL))
	if ie != nil {
		return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, ie.Error())
	}

	_, err = p.client.Get(context.Background(), p.path)
	// should get first, if not exist, set it
	if err != nil {
		if err != rpctypes.ErrKeyNotFound {
			return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, err.Error())
		}
		if _, err = p.client.Put(context.TODO(), p.path, p.src.Service, clientv3.WithLease(resp.ID)); err != nil {
			return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, err.Error())
		}
		return
	}

	// refresh set to true for not notifying the watcher
	if _, err = p.client.Put(context.Background(), p.path, p.src.Service, clientv3.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("grpclib: refresh service %q with ttl to clientv3 failed: %s", p.src.Name, err.Error())
	}
	return
}

// Revoke delete registered service from etcd
func (p *defaultServerRegister) Revoke() error {
	if p.client == nil {
		return nil
	}
	p.stopSignal <- true
	defer p.client.Close()
	_, err := p.client.Delete(context.Background(), p.path)
	return err
}

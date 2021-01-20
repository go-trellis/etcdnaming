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
	"math/rand"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"golang.org/x/net/context"
)

// defaultServerRegister default server register
type defaultServerRegister struct {
	src ServerRegisterConfig

	id string
	// prefix should start and end with no slash
	prefix string
	// invoke self-register with ticker
	ticker *time.Ticker
	// etcd key path
	path string

	client *clientv3.Client

	stopSignal chan bool
}

// NewDefaultServerRegister instance of server regitster
func NewDefaultServerRegister(c ServerRegisterConfig) ServerRegister {
	rand.Seed(time.Now().Unix())
	p := &defaultServerRegister{
		src: c,

		id:         fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(10000)),
		prefix:     fmt.Sprintf("/%s", schema),
		stopSignal: make(chan bool, 1),
		ticker:     time.NewTicker(c.Interval),
	}

	p.src.serverName = fmt.Sprintf("%s/%s", p.src.Name, p.src.Version)
	p.path = fmt.Sprintf("%s/%s/%s",
		p.prefix,
		p.src.serverName,
		p.src.Service)
	return p
}

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
				if p.src.RegistRetryTimes < 0 {
					continue
				}

				count++
				if p.src.RegistRetryTimes < count {
					panic(fmt.Errorf("%s regist into etcd failed times above: %d, %v",
						p.src.serverName, count, err))
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

	return
}

func (p *defaultServerRegister) regist() (err error) {
	// minimum lease TTL is ttl-second
	ctxGrant, cGrant := context.WithTimeout(context.TODO(), p.src.Interval)
	defer cGrant()
	resp, ie := p.client.Grant(ctxGrant, int64(p.src.TTL/time.Second))
	if ie != nil {
		return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, ie.Error())
	}

	ctxGet, cGet := context.WithTimeout(context.Background(), p.src.Interval)
	defer cGet()
	_, err = p.client.Get(ctxGet, p.path)
	// should get first, if not exist, set it
	if err != nil {
		if err != rpctypes.ErrKeyNotFound {
			return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, err.Error())
		}
		ctxPut, cPut := context.WithTimeout(context.TODO(), p.src.Interval)
		defer cPut()
		if _, err = p.client.Put(ctxPut, p.path, p.src.Service, clientv3.WithLease(resp.ID)); err != nil {
			return fmt.Errorf("grpclib: set service %q with ttl to clientv3 failed: %s", p.src.Name, err.Error())
		}
		return
	}

	// refresh set to true for not notifying the watcher
	ctxPut, cPut := context.WithTimeout(context.TODO(), p.src.Interval)
	defer cPut()
	if _, err = p.client.Put(ctxPut, p.path, p.src.Service, clientv3.WithLease(resp.ID)); err != nil {
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
	// refresh set to true for not notifying the watcher
	ctxDel, cDel := context.WithTimeout(context.TODO(), p.src.Interval)
	defer cDel()
	_, err := p.client.Delete(ctxDel, p.path)
	return err
}

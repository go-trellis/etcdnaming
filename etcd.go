// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// ServerRegister regist or revoke server interface
type ServerRegister interface {
	// Regist 注册服务
	Regist() error
	// Revoke 注销服务
	Revoke() error
}

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

// ServerRegisterConfig struct server regiter config
// name: server name
// target: etcd' client url, separate by ','
// serv: server address host:port
// interval: Rotation time to registe serv into etcd
// ttl: expired time, seconds
// registRetryTimes: allow failed to regist server and retry times
type ServerRegisterConfig struct {
	Name    string
	Target  string
	Service string // host & port

	TTL      int
	Interval time.Duration

	RegistRetryTimes int
}

const prefix = "etcd_naming"

// NewDefaultServerRegister instance of server regitster
func NewDefaultServerRegister(c ServerRegisterConfig) ServerRegister {
	rand.Seed(time.Now().Unix())
	p := &defaultServerRegister{
		src: c,

		id:     fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(10000)),
		prefix: prefix,

		stopSignal: make(chan bool, 1),
		ticker:     time.NewTicker(c.Interval),
	}
	p.path = fmt.Sprintf("/%s/%s/%s", p.prefix, p.src.Name, p.src.Service)

	return p
}

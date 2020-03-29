// GNU GPL v3 License
// Copyright (c) 2017 github.com:go-trellis

package etcdnaming

import "time"

// ServerRegister regist or revoke server interface
type ServerRegister interface {
	// Regist 注册服务
	Regist() error
	// Revoke 注销服务
	Revoke() error
}

// ServerRegisterConfig struct server regiter config
// name: server name
// target: etcd' client url, separate by ','
// serv: server address host:port
// interval: Rotation time to registe serv into etcd
// ttl: expired time, seconds
// registRetryTimes: allow failed to regist server and retry times; -1 alaways retry
type ServerRegisterConfig struct {
	Name             string
	Target           string
	Service          string // host & port
	Version          string // server's version
	TTL              time.Duration
	Interval         time.Duration
	RegistRetryTimes int

	serverName string // name & version
}

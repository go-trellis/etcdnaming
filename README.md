# etcdnaming
a tool which using ectd for some clients to find server when its started

## etcd & grpc

* [ETCD Link](https://github.com/coreos/etcd)  | vendor version：[v3.3.19](https://github.com/coreos/etcd/releases/tag/v3.3.19)
* [GRPC Link](https://github.com/grpc/grpc-go) | vendor version：[v1.26.0](https://github.com/grpc/grpc-go/releases/tag/v1.26.0)

## installation

* fisrt

github.com/coreos/go-systemd v22.0.0

```golang
go get github.com/coreos/go-systemd
```

* than: added this into go.mod

```
replace github.com/coreos/go-systemd => $GOPATH/src/github.com/coreos/go-systemd
```

* last:

```golang
go get github.com/go-trellis/etcdnaming
```

## Server

**[server example](go-server/main.go)**

```go
// ServerRegister regist or revoke server interface
type ServerRegister interface {
	// Regist 注册服务
	Regist() error
	// Revoke 注销服务
	Revoke() error
}
```

## Client

**[client example](go-client/main.go)**

```go
// NewResolver return resolver with service name
// serviceName server's name
// target: etcd client server address
func NewResolver(serviceName, target string, timeout time.Duration) error {
    return nil
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
```


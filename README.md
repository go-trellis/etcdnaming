# etcdnaming
a tool which using ectd for some clients to find server when its started

## etcd & grpc

* [ETCD Link](https://github.com/coreos/etcd)  | vendor version：[v3.3.19](https://github.com/coreos/etcd/releases/tag/v3.3.19)
* [GRPC Link](https://github.com/grpc/grpc-go) | vendor version：[v1.26.0](https://github.com/grpc/grpc-go/releases/tag/v1.26.0)

## installation

* fisrt

github.com/coreos/go-systemd v22.0.0

* than: added this into go.mod

```
replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
```

* last:

```golang
go get github.com/iTrellis/etcdnaming
```

## Server

**[server example](go-server/main.go)**


## Client

**[client example](go-client/main.go)**

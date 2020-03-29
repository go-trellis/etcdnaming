// GNU GPL v3 License
// Copyright (c) 2020 github.com:go-trellis

package etcdnaming

import (
	"time"

	"go.etcd.io/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcdnaming"

	// DialTarget 访问目标
	DialTarget = schema + ":///"
)

type etcdBuilder struct {
	opts BuilderOptions
}

type etcdResolver struct {
	opts BuilderOptions
	// grpc conn
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
	cc     resolver.ClientConn
	target resolver.Target

	key string

	rn chan struct{}
}

// BuilderOptions builder's configure
type BuilderOptions struct {
	Server     string
	Version    string
	Endpoint   string
	LooperTime time.Duration
}

// NewBuilder 获取builder
func NewBuilder(opt BuilderOptions) {
	b := &etcdBuilder{
		opts: opt,
	}

	resolver.Register(b)
}

func (p *etcdBuilder) Scheme() string {
	return schema
}

// Dial 拨号
func Dial(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), opts...)
}

// DialContext 带上下文的拨号
func DialContext(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	c, err := grpc.DialContext(ctx, DialTarget, opts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

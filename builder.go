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

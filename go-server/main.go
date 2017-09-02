// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-trellis/etcdnaming"
	"github.com/go-trellis/etcdnaming/go-server/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	serv = flag.String("serv", "127.0.0.1", "service listen address")
	port = flag.Int("port", 8001, "listening port")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

type server struct{}

func (*server) SayWorld(_ context.Context, req *proto.ReqSayWorld) (resp *proto.ReplySayWorld, err error) {
	fmt.Println("i am address", *serv, *port)
	fmt.Println("hello:", req.GetName())
	return &proto.ReplySayWorld{Message: "ok"}, nil
}

func main() {
	flag.Parse()

	ser := fmt.Sprintf("%s:%d", *serv, *port)

	lis, err := net.Listen("tcp", ser)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	register := etcdnaming.NewDefaultServerRegister(etcdnaming.ServerRegisterConfig{
		Name:    *serv,
		Target:  *reg,
		Service: ser,

		Interval: time.Second * 5,
		TTL:      8,

		RegistRetryTimes: 1,
	})

	err = register.Regist()
	if err != nil {
		log.Fatalf("regist hello: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterHelloServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

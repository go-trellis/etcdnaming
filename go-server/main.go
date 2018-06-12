// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-trellis/etcdnaming"
	"github.com/go-trellis/etcdnaming/go-server/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	name = flag.String("name", "hello", "service name")
	ver  = flag.String("ver", "v0", "server's version")
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
		Name:    *name,
		Target:  *reg,
		Service: ser,
		Version: *ver,

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

	// getting stop message and revoke etcd
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		message := <-ch
		log.Printf("receive signal '%v'", message)
		if err := register.Revoke(); err != nil {
			log.Printf("failure to revoke etcd: '%v'", err)
		} else {
			log.Printf("revoke etcd ok")
		}

		s.Stop()
		log.Println("stop grpc server")
	}()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

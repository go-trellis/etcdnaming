// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-trellis/etcdnaming"

	"github.com/go-trellis/etcdnaming/go-server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

var (
	name = flag.String("name", "hello", "service name")
	ver  = flag.String("ver", "v0", "server's version")
	cli  = flag.String("client", "1", "clinet name")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

func init() {
	if err := os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info"); err != nil {
		panic(err)
	}
}
func main() {
	flag.Parse()
	fmt.Println("connect to:", *name, *ver)

	etcdnaming.NewBuilder(*name, *ver, *reg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := etcdnaming.DialContext(ctx, grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))
	if err != nil {
		log.Printf("grpc dial: %s", err)
		return
	}
	client := proto.NewHelloClient(c)

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		resp, err := client.SayWorld(context.Background(), &proto.ReqSayWorld{Name: *cli + ": haha"})
		if err != nil {
			log.Println("aa:", err)
			time.Sleep(time.Second)
			continue
		}
		fmt.Printf("%v: Reply is %s\n", t, resp.Message)
	}

	// getting stop message and revoke etcd
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	message := <-ch
	log.Printf("receive signal '%v'", message)
	if err := c.Close(); err != nil {
		log.Printf("failure to close etcd client: '%v'", err)
	} else {
		log.Printf("close etcd ok")
	}
	time.Sleep(time.Second)
}

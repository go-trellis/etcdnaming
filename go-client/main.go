// GNU GPL v3 License

// Copyright (c) 2017 github.com:go-trellis

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-trellis/etcdnaming"
	"github.com/go-trellis/etcdnaming/go-sever/proto"
	uuid "github.com/satori/go.uuid"
)

var (
	name = flag.String("name", "hello", "service name")
	cli  = flag.String("client", "1", "clinet name")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()
	err := etcdnaming.NewResolver(*name, *reg, 10*time.Second)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		conn, ok := etcdnaming.GetResolverConn(*name)
		if !ok {
			fmt.Println("not get conn")
			return
		}
		resp, err := proto.NewHelloClient(conn).SayWorld(context.Background(),
			&proto.ReqSayWorld{Name: "world " + *cli + " :" + uuid.NewV4().String()})

		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
			continue
		}
		fmt.Println(err)
	}
}

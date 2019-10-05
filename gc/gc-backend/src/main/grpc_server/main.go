package main

import (
	"flag"
	"net"

	server "git.corp.angel-salon.com/gc/gc-backend/src"
	"git.corp.angel-salon.com/gc/proto"
	"google.golang.org/grpc"
)

var (
	flagPort = flag.String("port", ":12102", "Server binding port.")
)

func main() {
	flag.Parse()
	tcp, err := net.Listen("tcp", *flagPort)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	// t := server.newServer()

	srv := server.NewServer()
	proto.RegisterComputerServer(s, srv)
	s.Serve(tcp)
}

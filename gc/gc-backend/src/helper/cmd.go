package main

import (
	"context"
	"git.corp.angel-salon.com/gc/proto"
	"google.golang.org/grpc"
)

func main() {
	var client proto.ComputerClient
	conn, err := grpc.Dial(":12102", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client = proto.NewComputerClient(conn)
	_, err = client.CopyEnvironment(context.Background(), &proto.CopyEnvironmentRequest{
		Path:         "风险控制",
		OriginPrefix: "国税",
		NewPrefix:    "大企业",
	})
	if err != nil {
		panic(err)
	}
}

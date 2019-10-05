package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "git.corp.angel-salon.com/gc/proto"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:12102", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err);
	}

	defer conn.Close()

	client := pb.NewComputerClient(conn)
	res, err := client.QueryDB(context.Background(), &pb.Pipeline{
		Collections: []*pb.Collection{
			&pb.Collection{
				Name: "staff_query_1",
				Body: &pb.Collection_Projection{
					Projection: &pb.Projection{
						Input: "t_sto_staff",
						Fields: []*pb.Projection_Field{
							&pb.Projection_Field{
								Name: "id",
								Expression: &pb.Expression{
									Body: &pb.Expression_Field{
										Field: "id",
									},
								},
							},
							&pb.Projection_Field{
								Name: "name",
								Expression: &pb.Expression{
									Body: &pb.Expression_Field{
										Field: "name",
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to query: %v", err)
	}
	for _, c := range res.Columns {
		fmt.Print(c, " ")
	}
	fmt.Println()
	for _, r := range res.Row {
		for _, v := range r.Value {
			fmt.Print(v, " ")
		}
		fmt.Println()
	}
}
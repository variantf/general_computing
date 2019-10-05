package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "git.corp.angel-salon.com/gc/proto"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"

	"os"
)

var (
	app = kingpin.New("gc-client", "A general db middleware client")
)

func main() {
	conn, err := grpc.Dial("localhost:12102", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err);
	}

	defer conn.Close()
	client := pb.NewComputerClient(conn)

	app.Command("query", "test select functionality").Action(func(c *kingpin.ParseContext) error {
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
		return nil
	})

	app.Command("delete", "test delete functionality").Action(func(c *kingpin.ParseContext) error {
		_, err := client.DeleteDB(context.Background(), &pb.Filter{
			Input: "t_sto_staff",
			Expression: &pb.Expression{
				Body: &pb.Expression_Operation{
					Operation: &pb.Operation{
						Operator: pb.Operator_EQ,
						Operands: []*pb.Expression{
							&pb.Expression{
								Body: &pb.Expression_Field{
									Field: "id",
								},
							},
							&pb.Expression{
								Body: &pb.Expression_Literal{
									Literal: &pb.Literal{
										Body: &pb.Literal_StringValue{
											StringValue: "116",
										},
									},
								},
							},
						},
					},
				},
			},
		})

		return err
	})

	app.Command("insert", "test insert functionality").Action(func(c *kingpin.ParseContext) error {
		_, err:= client.InsertDB(context.Background(), &pb.NewRecords{
			Table: "t_sto_staff",
			Records: &pb.ResultSet{
				Columns: []string{"storeno","staffno", "name"},
				Row: []*pb.Result{
					&pb.Result{Value: []string{"111111", "222222", "jiayu"}},
					&pb.Result{Value: []string{"321321", "212234", "jiayu"}},
				},
			},
		})
		return err;
	})

	app.Command("update", "test update functionality").Action(func(c *kingpin.ParseContext) error {
		_, err := client.UpdateDB(context.Background(), &pb.UpdateRecords{
			Table: "t_sto_staff",
			Records: &pb.ResultSet{
				Columns: []string{"storeno","staffno", "name"},
				Row: []*pb.Result{&pb.Result{Value: []string{"9999", "8888", "jiayu"}}},
			},
			Condition: &pb.Expression{
				Body: &pb.Expression_Operation{
					Operation: &pb.Operation{
						Operator: pb.Operator_EQ,
						Operands: []*pb.Expression{
							&pb.Expression{
								Body: &pb.Expression_Field{
									Field: "id",
								},
							},
							&pb.Expression{
								Body: &pb.Expression_Literal{
									Literal: &pb.Literal{
										Body: &pb.Literal_StringValue{
											StringValue: "121",
										},
									},
								},
							},
						},
					},
				},
			},
		})

		return err
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
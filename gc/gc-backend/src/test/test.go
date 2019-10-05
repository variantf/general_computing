package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "gitlab.com/jsq/general_computing/src/proto"
	"google.golang.org/grpc"
	"io/ioutil"
)

func main() {
	a := pb.Table{
		Path: "/风控/国税",
		Fields: map[string]*pb.Table_Field{
			"A": &pb.Table_Field{
				Name: "A",
				Type: pb.Type_FLOAT,
			},
			"C": &pb.Table_Field{
				Name: "C",
				Type: pb.Type_FLOAT,
			},
		},
		Pks: []string{"A"},
	}

	conn, err := grpc.Dial("localhost:12100", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewDataManagerClient(conn)
	byt, _ := ioutil.ReadFile("keys.json")
	var data []string
	fields := map[string]*pb.Table_Field{}
	json.Unmarshal(byt, &data)
	for _, fieldName := range data {
		fields[fieldName] = &pb.Table_Field{
			Name: fieldName,
			Type: pb.Type_FLOAT,
		}
	}
	fields["Company"] = &pb.Table_Field{
		Name: "Company",
		Type: pb.Type_STRING,
	}
	fields["Year"] = &pb.Table_Field{
		Name: "Year",
		Type: pb.Type_FLOAT,
	}
	fields["本单位申报"] = &pb.Table_Field{
		Name: "本单位申报",
		Type: pb.Type_BOOLEAN,
	}

	response, err := client.UpdateTable(context.Background(), &pb.UpdateTableRequest{
		Table: &a,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(response)

	// _, err = client.UpdateData(context.Background(), &pb.UpdateDataRequest{
	// 	Entity: &pb.Entity{
	// 		Path: "/风控/国税",
	// 		Fields: []*pb.Entity_Field{
	// 			&pb.Entity_Field{
	// 				Name: "Company",
	// 				Value: &pb.Literal{
	// 					Body: &pb.Literal_StringValue{StringValue: "测试公司"},
	// 				},
	// 			},
	// 			&pb.Entity_Field{
	// 				Name: "Year",
	// 				Value: &pb.Literal{
	// 					Body: &pb.Literal_FloatValue{FloatValue: 2015},
	// 				},
	// 			},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// create_response, err := client.CreateFormula(context.Background(), &pb.CreateFormulaRequest{
	// 	Path: "/风控/国税",
	// 	Name: "测试公式1",
	// })
	// fmt.Println(create_response)

	// formulas, err := client.FetchFormula(context.Background(), &pb.FetchFormulaRequest{
	// 	PathPattern: "/风控/国税",
	// 	Type:        pb.Formula_SAVE_ONLY,
	// 	Name:        "测试公式1",
	// })
	// if err != nil {
	// 	fmt.Println("FetchFormula")
	// 	panic(err)
	// }
	// formula := formulas.Formulas[0]
	// pipeline := formula.Pipeline
	// pipeline.Collections = []*pb.Collection{pipeline.Collections[0]}
	// pipeline.Collections = append(pipeline.Collections, &pb.Collection{
	// 	Name: "投影",
	// 	Body: &pb.Collection_Projection{
	// 		Projection: &pb.Projection{
	// 			Input: "数据源",
	// 			Fields: []*pb.Projection_Field{
	// 				&pb.Projection_Field{
	// 					Name: "常量",
	// 					Expression: &pb.Expression{
	// 						Body: &pb.Expression_Literal{
	// 							Literal: &pb.Literal{
	// 								Body: &pb.Literal_FloatValue{
	// 									FloatValue: 1.234,
	// 								},
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// })

	// _, err = client.UpdateFormula(context.Background(), &pb.UpdateFormulaRequest{
	// 	Formula: &pb.Formula{
	// 		Path:     "/风控/国税",
	// 		Name:     "测试公式1",
	// 		Type:     pb.Formula_SAVE_ONLY,
	// 		Pipeline: pipeline,
	// 	},
	// })

	// for _, col := range pipeline.Collections {
	// 	fmt.Println(col.Name)
	// }

	// if err != nil {
	// 	fmt.Println("UpdateFormula")
	// 	panic(err)
	// }

	// dbg_response, err := client.DebugFormula(context.Background(), &pb.DebugFormulaRequest{
	// 	Path: "/风控/国税",
	// 	Name: "测试公式1",
	// })

	// if err != nil {
	// 	fmt.Println("DebugFormula")
	// 	panic(err)
	// }
	// fmt.Print(dbg_response)
}

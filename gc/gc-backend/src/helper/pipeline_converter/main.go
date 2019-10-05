package main

import (
	"flag"
	"fmt"
	"strings"

	pb "gitlab.com/jsq/general_computing/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	flagStart = flag.Int("start", 0, "")
	flagLimit = flag.Int("limit", 1, "")

	tableMetadata map[string]map[string]*pb.TableMetadata
)

func ConvertPipeline(pipeline *pb.Pipeline) (*pb.Pipeline, error) {
	isInput := make(map[string]bool)
	for _, colle := range pipeline.Collections {
		switch colle.Body.(type) {
		case *pb.Collection_Input:
			isInput[colle.Name] = true
		default:
			isInput[colle.Name] = false
		}
		if strings.HasSuffix(colle.Name, "_日期提取表") {
			return nil, nil
		}
	}
	fmt.Println("Input tables: ", isInput)
	var collections []*pb.Collection
	for _, colle := range pipeline.Collections {
		switch body := colle.Body.(type) {
		case *pb.Collection_Input:
			collections = append(collections, colle)

			table := tableMetadata[pipeline.Path][body.Input.MetaName]
			var fields []*pb.Projection_Field
			for _, field := range table.Fields {
				rune_name := []rune(field.Name)
				if strings.HasSuffix(field.Name, "日期") || strings.HasSuffix(field.Name, "年月") || strings.HasPrefix(field.Name, "所属期") {
					var prefix string
					if strings.HasPrefix(field.Name, "所属期") {
						prefix = string(rune_name)
					} else {
						prefix = string(rune_name[:len(rune_name)-2])
					}
					fields = append(fields, &pb.Projection_Field{
						Name: prefix + "年",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_YEAR,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					}, &pb.Projection_Field{
						Name: prefix + "月",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_MONTH,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					}, &pb.Projection_Field{
						Name: prefix + "日",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_DAY,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					})
				} else if strings.HasSuffix(field.Name, "RQ") {
					prefix := string(rune_name)
					fields = append(fields, &pb.Projection_Field{
						Name: prefix + "-Year",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_YEAR,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					}, &pb.Projection_Field{
						Name: prefix + "-Month",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_MONTH,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					}, &pb.Projection_Field{
						Name: prefix + "-Day",
						Expression: &pb.Expression{
							Body: &pb.Expression_Operation{
								Operation: &pb.Operation{
									Operator: pb.Operator_DAY,
									Operands: []*pb.Expression{&pb.Expression{Body: &pb.Expression_Field{Field: field.Name}}},
								},
							},
						},
					})
				} else {
					fields = append(fields, &pb.Projection_Field{
						Name: field.Name,
						Expression: &pb.Expression{
							Body: &pb.Expression_Field{Field: field.Name},
						},
					})
				}
			}

			collections = append(collections, &pb.Collection{
				Name: colle.Name + "_日期提取表",
				Body: &pb.Collection_Projection{
					Projection: &pb.Projection{
						Input:  colle.Name,
						Fields: fields,
					},
				},
			})

		case *pb.Collection_Filter:
			if isInput[body.Filter.Input] {
				collections = append(collections, &pb.Collection{
					Name: colle.Name,
					Body: &pb.Collection_Filter{
						Filter: &pb.Filter{
							Input:      body.Filter.Input + "_日期提取表",
							Expression: body.Filter.Expression,
						},
					},
				})
			} else {
				collections = append(collections, colle)
			}
		case *pb.Collection_Projection:
			if isInput[body.Projection.Input] {
				collections = append(collections, &pb.Collection{
					Name: colle.Name,
					Body: &pb.Collection_Projection{
						Projection: &pb.Projection{
							Input:  body.Projection.Input + "_日期提取表",
							Fields: body.Projection.Fields,
						},
					},
				})
			} else {
				collections = append(collections, colle)
			}
		case *pb.Collection_Join:
			var leftInput, rightInput string
			if isInput[body.Join.LeftInput] {
				leftInput = body.Join.LeftInput + "_日期提取表"
			} else {
				leftInput = body.Join.LeftInput
			}
			if isInput[body.Join.RightInput] {
				rightInput = body.Join.RightInput + "_日期提取表"
			} else {
				rightInput = body.Join.RightInput
			}
			collections = append(collections, &pb.Collection{
				Name: colle.Name,
				Body: &pb.Collection_Join{
					Join: &pb.Join{
						LeftInput:   leftInput,
						RightInput:  rightInput,
						Conditions:  body.Join.Conditions,
						Method:      body.Join.Method,
						LeftFields:  body.Join.LeftFields,
						RightFields: body.Join.RightFields,
					},
				},
			})
		case *pb.Collection_Group:
			if isInput[body.Group.Input] {
				collections = append(collections, &pb.Collection{
					Name: colle.Name,
					Body: &pb.Collection_Group{
						Group: &pb.Group{
							Input:  body.Group.Input + "_日期提取表",
							Keys:   body.Group.Keys,
							Fields: body.Group.Fields,
						},
					},
				})
			} else {
				collections = append(collections, colle)
			}
		}
	}
	return &pb.Pipeline{
		Collections:    collections,
		Path:           pipeline.Path,
		Name:           pipeline.Name,
		ResultMetaName: pipeline.ResultMetaName,
	}, nil
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial("localhost:12100", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewComputerClient(conn)

	resp, err := client.FetchPipelineList(context.Background(), &pb.Empty{})
	if err != nil {
		panic(err)
	}

	tableMetadata = make(map[string]map[string]*pb.TableMetadata)
	for _, pipeline := range resp.Pipelines {
		if _, ok := tableMetadata[pipeline.Path]; ok {
			continue
		}
		tables, err := client.TableMetadataList(context.Background(), &pb.TableMetadataRequest{Path: pipeline.Path})
		if err != nil {
			panic(err)
		}
		tableMetadata[pipeline.Path] = make(map[string]*pb.TableMetadata)
		for _, table := range tables.TableMetadata {
			tableMetadata[pipeline.Path][table.Name] = table
		}
	}

	var result []*pb.Pipeline
	for i, pipeline := range resp.Pipelines {
		if i < *flagStart || i >= *flagStart+*flagLimit {
			continue
		}
		pipeline, err = client.FetchPipeline(context.Background(), &pb.PathName{Path: pipeline.Path, Name: pipeline.Name})
		if err != nil {
			panic(err)
		}
		// fmt.Println("Before convert: ", pipeline)
		pipeline, err = ConvertPipeline(pipeline)
		if err != nil {
			panic(err)
		}
		if pipeline != nil {
			result = append(result, pipeline)
			fmt.Println("After convert: ", pipeline)
		}
	}
	for _, pipeline := range result {
		_, err := client.UpdatePipeline(context.Background(), pipeline)
		if err != nil {
			panic(err)
		}
	}
}

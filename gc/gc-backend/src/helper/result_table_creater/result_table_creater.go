package main

import (
	"fmt"

	pb "git.corp.angel-salon.com/gc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var result_info = map[string][]*pb.TableMetadata_Field{
	"税款输出模板": []*pb.TableMetadata_Field{
		&pb.TableMetadata_Field{
			Name: "公司名称",
			Type: pb.Type_STRING,
		},
		&pb.TableMetadata_Field{
			Name: "年份",
			Type: pb.Type_FLOAT,
		},
		&pb.TableMetadata_Field{
			Name: "月份",
			Type: pb.Type_FLOAT,
		},
		&pb.TableMetadata_Field{
			Name: "描述",
			Type: pb.Type_STRING,
		},
		&pb.TableMetadata_Field{
			Name: "税款",
			Type: pb.Type_FLOAT,
		},
		&pb.TableMetadata_Field{
			Name: "税种",
			Type: pb.Type_STRING,
		},
	},
	"风险输出模板": []*pb.TableMetadata_Field{
		&pb.TableMetadata_Field{
			Name: "公司名称",
			Type: pb.Type_STRING,
		},
		&pb.TableMetadata_Field{
			Name: "年份",
			Type: pb.Type_FLOAT,
		},
		&pb.TableMetadata_Field{
			Name: "月份",
			Type: pb.Type_FLOAT,
		},
		&pb.TableMetadata_Field{
			Name: "描述",
			Type: pb.Type_STRING,
		},
		&pb.TableMetadata_Field{
			Name: "类型",
			Type: pb.Type_STRING,
		},
	},
}

var db_info = map[string]string{
	"税款输出模板": "RESULT_TAX",
	"风险输出模板": "RESULT_RISK",
}

func main() {
	// 插入Metadata并构造表结构
	conn, err := grpc.Dial("192.168.44.13:12100", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewDataManagerClient(conn)
	for metaName, metaFields := range result_info {
		updateTable := pb.UpdateTableRequest{
			Table: &pb.TableMetadata{
				Name:   metaName,
				Path:   "风控",
				Fields: metaFields,
			},
			Remove: true,
		}
		_, err = client.UpdateTableMeta(context.Background(), &updateTable)
		if err != nil {
			fmt.Println("error: ", err)
		}

		_, err = client.CreateDatabaseTable(context.Background(), &pb.CreateDatabaseTableRequest{
			MetaName: metaName,
			MetaPath: "风控",
			DbName:   db_info[metaName],
		})
		if err != nil {
			fmt.Println("error: ", err)
		}
	}
}

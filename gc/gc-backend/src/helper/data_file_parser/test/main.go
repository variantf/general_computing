package main

import (
	"fmt"
	pb "git.corp.angel-salon.com/gc/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func main() {
	app := kingpin.New("client", "UMA debug tool")
	server := app.Flag("server", "Server address").Default(":12102").String()

	var client pb.DataManagerClient
	app.Action(func(*kingpin.ParseContext) error {
		conn, err := grpc.Dial(*server, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		client = pb.NewDataManagerClient(conn)
		return nil
	})
	{
		auth := app.Command("FetchDatabaseTable", "FetchDatabaseTable with db_name")
		db_name := auth.Arg("db_name", "db_name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.FetchDatabaseTable(context.Background(), &pb.FetchDatabaseTableRequest{DbName: *db_name})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			log.Fatal(response.Fields[0].Alias)
			return nil
		})
	}
	{
		auth := app.Command("UpsertDatabaseTable", "UpsertDatabaseTable with many")
		// path := auth.Arg("path", "Path").Required().String()
		// fmt.Println(path)
		//upsertdatbase
		meta_path := auth.Arg("meta_path", "meta_path").Required().String()
		meta_name := auth.Arg("meta_name", "meta_name").Required().String()
		db_name := auth.Arg("db_name", "db_name").Required().String()

		std_name := auth.Arg("std_name", "std_name").Required().String()
		Alias := auth.Arg("Alias", "Alias").Required().Strings()

		auth.Action(func(*kingpin.ParseContext) error {
			// response, err := client.FetchTableMetadata(context.Background(), &pb.PathName{Path: *path, Name: *name})
			_, err := client.UpsertDatabaseTable(context.Background(), &pb.DatabaseTable{
				MetaPath: *meta_path,
				MetaName: *meta_name,
				DbName:   *db_name,
				Fields: []*pb.DatabaseTable_FieldMapping{
					&pb.DatabaseTable_FieldMapping{StdName: *std_name, Alias: *Alias},
				},
			})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			return nil
		})
	}
	{
		auth := app.Command("AnalyzeDataFile", "AnalyzeDataFile with AnalyzeDataFileRequest")
		db_name := auth.Arg("db_name", "db_name").Required().String()
		file_name := auth.Arg("file_name", "file_name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.AnalyzeDataFile(context.Background(), &pb.AnalyzeDataFileRequest{DbName: *db_name, FileName: *file_name})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			log.Fatal(response.Message)
			return nil
		})
	}
	{
		auth := app.Command("DeleteTableMetadata", "DeleteTableMetadata with pathname")
		name := auth.Arg("name", "name").Required().String()
		path := auth.Arg("path", "path").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.DeleteTableMetadata(context.Background(), &pb.PathName{Path: *path, Name: *name})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			log.Fatal(response)
			return nil
		})
	}
	{
		auth := app.Command("DeleteDatabaseTable", "DeleteDatabaseTable with db_name")
		db_name := auth.Arg("db_name", "db_name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.DeleteDatabaseTable(context.Background(), &pb.DeleteDatabaseTableRequest{DbName: *db_name})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			log.Fatal(response)
			return nil
		})
	}

	{
		auth := app.Command("LoadDataFromDB", "LoadDataFromDB with LoadDataFromDBRequest")
		db_conn_name := auth.Arg("db_conn_name", "db_conn_name").Required().String()
		group := auth.Arg("group", "group").Required().String()
		table := auth.Arg("table", "table").Required().String()
		target_db := auth.Arg("target_db", "target_db").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.LoadDataFromDB(context.Background(), &pb.LoadDataFromDBRequest{
				DbConnName: *db_conn_name,
				Group:      *group,
				Table:      *table,
				TargetDb:   *target_db})
			fmt.Println("resp", &pb.AnalyzeDataFileResponse{})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			log.Info(response.Message)
			return nil
		})
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

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

	var client pb.ComputerClient
	app.Action(func(*kingpin.ParseContext) error {
		conn, err := grpc.Dial(*server, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		client = pb.NewComputerClient(conn)
		return nil
	})

	{
		auth := app.Command("InsertOracleMsg", "InsertOracleMsg with many")
		name := auth.Arg("name", "name").Required().String()
		usrname := auth.Arg("usrname", "usrname").Required().String()
		password := auth.Arg("password", "password").Required().String()
		ipadd := auth.Arg("ipadd", "ipadd").Required().String()
		port := auth.Arg("port", "port").Required().String()
		instance := auth.Arg("instance", "instance").Required().String()
		databasename := auth.Arg("databasename", "databasename").Required().String()

		auth.Action(func(*kingpin.ParseContext) error {
			res, err := client.InsertOracleMsg(context.Background(), &pb.OracleMsg{
				Name:         *name,
				Usrname:      *usrname,
				Password:     *password,
				Ipadd:        *ipadd,
				Port:         *port,
				Instance:     *instance,
				Databasename: *databasename,
			})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			} else {
				fmt.Println(res)
			}
			return nil
		})
	}

	{
		auth := app.Command("OracleMsgList", "OracleMsgList with nothing")
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.OracleMsgList(context.Background(), &pb.Empty{})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			// for _, v := range response.Oraclemsgs {
			log.Fatal(response.Oraclemsgs[0])
			// }
			return err

		})
	}

	{
		auth := app.Command("DeleteOracleMsg", "DeleteOracleMsg with ")
		name := auth.Arg("name", "name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			_, err := client.DeleteOracleMsg(context.Background(), &pb.DeleteOracleMsgResquest{Name: *name})
			if err != nil {
				fmt.Println(err)
			}
			return err

		})
	}
	{
		auth := app.Command("FetchPipelineList", "FetchPipelineList with empty")
		// db_name := auth.Arg("db_name", "db_name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.FetchPipelineList(context.Background(), &pb.Empty{})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			fmt.Println(response)
			return nil
		})
	}

	{
		auth := app.Command("DuplicateTaskgs", "DuplicateTaskgs with path")
		path := auth.Arg("path", "path").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.DuplicateTaskgs(context.Background(),
				&pb.DuplicateTaskgsRequest{
					Path: *path,
				})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			fmt.Println(response)
			return nil
		})
	}
	{
		auth := app.Command("FetchTestResult", "FetchTestResult with path name")
		path := auth.Arg("path", "path").Required().String()
		name := auth.Arg("name", "name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.FetchTestResult(context.Background(),
				&pb.PathName{
					Path: *path,
					Name: *name,
				})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			fmt.Println(response)
			return nil
		})
	}
	{
		auth := app.Command("FetchRunTime", "FetchRunTime with path name")
		path := auth.Arg("path", "path").Required().String()
		name := auth.Arg("name", "name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.FetchRunTime(context.Background(),
				&pb.PathName{
					Path: *path,
					Name: *name,
				})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			fmt.Println(response)
			return nil
		})
	}

	{
		auth := app.Command("DebugTask", "DebugTask with path name")
		path := auth.Arg("path", "path").Required().String()
		name := auth.Arg("name", "name").Required().String()
		auth.Action(func(*kingpin.ParseContext) error {
			response, err := client.DebugTask(context.Background(),
				&pb.PathName{
					Path: *path,
					Name: *name,
				})
			if err != nil {
				fmt.Println(err)
				// log.Fatal(err.Error())
			}
			fmt.Println(response)
			return nil
		})
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

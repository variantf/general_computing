package server

import (
	"database/sql"
	"fmt"

	pb "git.corp.angel-salon.com/gc/proto"
	"golang.org/x/net/context"
	//_ "github.com/mattn/go-oci8"
)

type SQLJob struct {
	Path    string
	Name    string
	Testing bool
}

// Server implements pb.ComputerServer.
type Server struct {
	jobQueue chan SQLJob
}

func NewServer() *Server {
	return &Server{
		jobQueue: make(chan SQLJob, 1024),
	}
}

func (s *Server) DeleteDB(ctx context.Context, req *pb.Filter) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) InsertDB(ctx context.Context, req *pb.NewRecords) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) UpdateDB(ctx context.Context, req *pb.ResultSet) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) QueryDB(ctx context.Context, req *pb.Pipeline) (*pb.ResultSet, error) {
	sqlQry, err := s.CompileTask(&pb.Task{}, req)
	for _, c := range req.Collections {
		fmt.Println("collection name: ", c.Name)
	}
	if err != nil {
		return nil, err
	}

	lastName := req.Collections[len(req.Collections)-1].Name
	sqlQry = sqlQry + "SELECT * FROM `" + lastName + "` ;"

	fmt.Println(sqlQry)
	rows, err := QuerySQL(sqlQry)
	if err != nil {
		return nil, err
	}

	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	resultSet := &pb.ResultSet{}
	resultSet.Columns = colNames
	row_data := make([]sql.NullString, len(colNames))
	row_ptrs := make([]interface{}, len(colNames))
	for i := range row_data {
		row_ptrs[i] = &row_data[i]
	}
	for rows.Next() {
		err = rows.Scan(row_ptrs...)
		if err != nil {
			return nil, err
		}
		result := &pb.Result{}
		for _, data := range row_data {
			if data.Valid {
				result.Value = append(result.Value, data.String)
			} else {
				result.Value = append(result.Value, "")
			}
		}
		resultSet.Row = append(resultSet.Row, result)
	}

	return resultSet, nil
}

package server

import (
	"database/sql"
	"fmt"

	pb "git.corp.angel-salon.com/gc/proto"
	"golang.org/x/net/context"
	"strings"
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
	sql := "DELETE FROM " + quoteIdentifier(req.Input) + " WHERE " + ParseExpression(req.Expression)
	fmt.Println("Exec SQL: ", sql)
	err := ExecSQL(sql)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *Server) InsertDB(ctx context.Context, req *pb.NewRecords) (*pb.Empty, error) {
	quotedColumns := []string{}
	oneRowPlaceholders := []string{}
	for _, col := range req.Records.Columns {
		quotedColumns = append(quotedColumns, quoteIdentifier(col))
		oneRowPlaceholders = append(oneRowPlaceholders, "?")
	}

	valuesPlaceholer := []string{}
	values := []interface{}{}
	for _, row := range req.Records.Row {
		for _, col := range row.Value {
			values = append(values, col)
		}
		valuesPlaceholer = append(valuesPlaceholer, "(" + strings.Join(oneRowPlaceholders, ", ") + ")")
	}
	
	sql := "INSERT INTO " + quoteIdentifier(req.Table) + 
		"(" + strings.Join(quotedColumns, ",") + ") VALUES" + strings.Join(valuesPlaceholer, ", ");

    fmt.Println(sql, values)
	err := ExecSQL(sql, values...)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *Server) UpdateDB(ctx context.Context, req *pb.UpdateRecords) (*pb.Empty, error) {
	updateColumns := []string{}
	if len(req.Records.Row) != 1 {
		return nil, fmt.Errorf("更新的数据必须为1行")
	}

	if len(req.Records.Columns) != len(req.Records.Row[0].Value) {
		return nil, fmt.Errorf("更新的列数量和值数量不一致")
	}

	values := []interface{}{}
	for idx, col := range req.Records.Columns {
		values = append(values, req.Records.Row[0].Value[idx])
		updateColumns = append(updateColumns, quoteIdentifier(col) + " = ?")
	}
	sql := "UPDATE " + quoteIdentifier(req.Table) + " SET " + strings.Join(updateColumns, ", ") + " WHERE " + ParseExpression(req.Condition)
	fmt.Println(sql, values)
	err := ExecSQL(sql, values...)
	if err != nil {
		return nil, err
	}
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

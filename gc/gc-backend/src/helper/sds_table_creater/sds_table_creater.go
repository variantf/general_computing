//+build !ZCZJTXQKB

package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/tealeg/xlsx"
	pb "gitlab.com/jsq/general_computing/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	flagDirectory = flag.String("d", "金三格式/", "xlsx directory")
	flagFilename  = flag.String("f", "", "file name")
	flagEWCOL     = flag.String("e", "EWBHXH", "ewb column title")
)

var mapping map[string]interface{}

func main() {
	flag.Parse()
	jsonfile, _ := ioutil.ReadFile("sds_mapping.json")
	// jsonfile, _ := ioutil.ReadFile("ZCZJTXQKB.json")
	err := json.Unmarshal(jsonfile, &mapping)
	if err != nil {
		panic(err)
	}
	if *flagFilename != "" {
		update_one_table(*flagFilename)
		return
	}
	files, _ := ioutil.ReadDir(*flagDirectory)
	for _, f := range files {
		// fmt.Println(f.Name())
		update_one_table(f.Name()[:len(f.Name())-5])
	}
}

func update_one_table(name string) {
	map_val, ok := mapping[strings.ToUpper(name)]
	if !ok {
		fmt.Println("no such mapping: ", name)
		return
	}
	table, _ := map_val.(map[string]interface{})
	// fmt.Println(name)
	excel, err := xlsx.OpenFile(*flagDirectory + name + ".xlsx")
	if err != nil {
		panic(err)
	}
	name, _ = table["name"].(string)
	// fmt.Println(name)
	tfs, _ := table["fields"].(map[string]interface{})
	// fmt.Println("load done")
	// meta := pb.TableMetadata{}
	// 从标题行获取列名
	fields := []*pb.TableMetadata_Field{}
	headRow := excel.Sheets[0].Rows[0]
	two_d := false //是否二维表
	two_d_cnt := 0 //二维坐标最大值
	for _, cell := range headRow.Cells {
		val, _ := cell.String()
		val = strings.TrimLeft(val, " ")
		val = strings.TrimRight(val, " ")
		if val == "SBUUID" || val == "NSRMC" {
			fields = append(fields, &pb.TableMetadata_Field{Name: val, Type: pb.Type_STRING})
			continue
		}
		val, ok := tfs[val].(string)
		if !ok {
			fields = append(fields, &pb.TableMetadata_Field{Name: "", Type: pb.Type_STRING})
		} else {
			fields = append(fields, &pb.TableMetadata_Field{Name: val, Type: pb.Type_FLOAT})
		}
		if val == *flagEWCOL {
			two_d = true
		}
	}
	// fmt.Println("head done")

	// 分析列类型
	for _, sheet := range excel.Sheets {
		for _, row := range sheet.Rows[1:] {
			for idx, cell := range row.Cells {
				if idx >= len(fields) {
					break
				}
				if fields[idx].Type == pb.Type_STRING {
					continue
				}
				if fields[idx].Name == *flagEWCOL {
					num, err := cell.Int()
					if err == nil && num > two_d_cnt {
						two_d_cnt = num
					}
					continue
				}
				if strings.HasSuffix(fields[idx].Name, "_DM") {
					fields[idx].Type = pb.Type_STRING
					continue
				}
				num, err := cell.Float()
				if err != nil || num > 1e15 {
					val, err := cell.String()
					if val == "" {
						continue
					}
					if err != nil {
						fmt.Println("warning", err)
					}
					if val != "" {
						fields[idx].Type = pb.Type_STRING
					}
				}
			}
		}
	}

	// fmt.Println(two_d, two_d_cnt)
	// 转化为一维的列名
	table_fields := []*pb.TableMetadata_Field{}
	for _, field := range fields {
		// rune_name := []rune(field.Name)
		// 时间列
		if strings.HasSuffix(field.Name, "日期") {
			table_fields = append(table_fields, &pb.TableMetadata_Field{Name: field.Name, Type: pb.Type_FLOAT})
			// prefix := string(rune_name[:len(rune_name)-2])
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "年", Type: pb.Type_FLOAT})
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "月", Type: pb.Type_FLOAT})
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "日", Type: pb.Type_FLOAT})
		} else if strings.HasSuffix(field.Name, "年月") {
			table_fields = append(table_fields, &pb.TableMetadata_Field{Name: field.Name, Type: pb.Type_FLOAT})
			// prefix := string(rune_name[:len(rune_name)-2])
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "年", Type: pb.Type_FLOAT})
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "月", Type: pb.Type_FLOAT})
		} else if strings.HasSuffix(field.Name, "RQ") {
			table_fields = append(table_fields, &pb.TableMetadata_Field{Name: field.Name, Type: pb.Type_FLOAT})
			// prefix := string(rune_name[:len(rune_name)])
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "-Year", Type: pb.Type_FLOAT})
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "-Month", Type: pb.Type_FLOAT})
			// table_fields = append(table_fields, &pb.TableMetadata_Field{Name: prefix + "-Day", Type: pb.Type_FLOAT})
		} else if field.Name != "" && (field.Type == pb.Type_STRING && field.Name != *flagEWCOL || !two_d) {
			// 一维表，或者二维表的字符串列
			table_fields = append(table_fields, &pb.TableMetadata_Field{Name: field.Name, Type: field.Type})
		} else if field.Name != "" && field.Type == pb.Type_FLOAT && field.Name != *flagEWCOL && two_d {
			// 二维表的实数列
			for i := 1; i <= two_d_cnt; i++ {
				table_fields = append(table_fields, &pb.TableMetadata_Field{Name: fmt.Sprintf("%s[%d]", field.Name, i), Type: pb.Type_FLOAT})
			}
		}
	}
	field_id := make(map[string]int)
	for idx, field := range table_fields {
		if _, ok := field_id[field.Name]; ok {
			fmt.Println("warning: same column name ", field.Name)
			return
		}
		field_id[field.Name] = idx
		fmt.Println(field.Name, field.Type)
	}

	// 插入Metadata并构造表结构
	conn, err := grpc.Dial("192.168.44.13:12100", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewDataManagerClient(conn)
	updateTable := pb.UpdateTableRequest{
		Table: &pb.TableMetadata{
			Name:   name,
			Path:   "风控",
			Fields: table_fields,
		},
		Remove: true,
	}
	_, err = client.UpdateTableMeta(context.Background(), &updateTable)
	if err != nil {
		panic(err)
	}
	_, err = client.CreateDatabaseTable(context.Background(), &pb.CreateDatabaseTableRequest{
		MetaName: name,
		MetaPath: "风控",
		DbName:   "国税_" + name,
	})
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", "postgres://postgres:AppStore321@192.168.44.13/general_computing?sslmode=require")
	if err != nil {
		panic(err)
	}

	txn, err := db.Begin()
	if err != nil {
		panic(err)
	}

	field_list := []string{}
	for _, field := range table_fields {
		field_list = append(field_list, field.Name)
	}

	stmt, err := txn.Prepare(pq.CopyIn("国税_"+name, field_list...))
	if err != nil {
		panic(err)
	}

	// 插入实际数据
	first := true
	company := ""
	values := []interface{}{}
	for _, sheet := range excel.Sheets {
		for _, row := range sheet.Rows[1:] {
			first = true
			EWBLXH := 0
			for idx, cell := range row.Cells {
				if idx >= len(fields) {
					break
				}
				if fields[idx].Name == "NSRMC" {
					str_val, _ := cell.String()
					if str_val == company && two_d {
						first = false
					}
					company = str_val
				} else if fields[idx].Name == *flagEWCOL {
					int_val, _ := cell.Int()
					EWBLXH = int_val
				}
			}
			if first {
				if two_d && len(values) == len(table_fields) {
					/*
						for idx, val := range values {
							if val != nil {
								fmt.Print(table_fields[idx].Name, " | ", val, " | ")
							}
						}
						fmt.Println()
					*/
					_, err = stmt.Exec(values...)
					if err != nil {
						panic(err)
					}
				}
				values = []interface{}{}
			}
			// fmt.Println(_i, _j, company, EWBLXH, first)
			for idx, cell := range row.Cells {
				if idx >= len(fields) {
					break
				}
				if strings.HasSuffix(fields[idx].Name, "日期") || strings.HasSuffix(fields[idx].Name, "RQ") {
					excel_time, err := cell.Float()
					if err != nil {
						// fmt.Println("warning date", err)
						if first {
							values = append(values, nil)
							// values = append(values, nil, nil, nil)
						}
						continue
					}
					time := xlsx.TimeFromExcelTime(excel_time, false)
					// year, month, date := time.Date()
					if first {
						values = append(values, time.Unix())
						// values = append(values, float64(year))
						// values = append(values, float64(month))
						// values = append(values, float64(date))
					}
				} else if strings.HasSuffix(fields[idx].Name, "年月") {
					val, err := cell.Float()
					if err != nil {
						// fmt.Println("warning date", err)
						if first {
							values = append(values, nil)
							// values = append(values, nil, nil)
						}
						continue
					}
					date := int32(val)
					year := date / 100
					month := date % 100
					if first {
						layout := "2006-01-02T15:04:05.000Z"
						str := fmt.Sprintf("%d-%02d-01T00:00:00.000Z", year, month)
						t, err := time.Parse(layout, str)
						if err != nil {
							fmt.Println(err)
							values = append(values, nil)
						} else {
							values = append(values, t.Unix())
						}
						// values = append(values, year)
						// values = append(values, month)
					}
				} else if fields[idx].Type == pb.Type_STRING && fields[idx].Name != "" && fields[idx].Name != *flagEWCOL {
					str_val, err := cell.String()
					if err != nil || str_val == "" {
						if err != nil {
							fmt.Println("warning string", err)
						}
						if first {
							values = append(values, nil)
						}
					} else {
						if first {
							values = append(values, str_val)
						}
					}
				} else if fields[idx].Type == pb.Type_FLOAT && fields[idx].Name != "" && fields[idx].Name != *flagEWCOL {
					if first && two_d {
						for i := 1; i <= two_d_cnt; i++ {
							values = append(values, nil)
						}
					}
					float_val, err := cell.Float()
					if err != nil {
						if !two_d {
							values = append(values, nil)
						}
					} else {
						if two_d {
							values[field_id[fmt.Sprintf("%s[%d]", fields[idx].Name, EWBLXH)]] = float_val
						} else {
							values = append(values, float_val)
						}
					}
				}
			}
			for len(values) < len(table_fields) {
				values = append(values, nil)
			}
			if len(values) != len(table_fields) {
				fmt.Println(len(values))
				fmt.Println(len(table_fields))
				panic("mismatch size")
			}
			if !two_d {
				_, err = stmt.Exec(values...)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	if two_d && len(values) == len(table_fields) {
		_, err = stmt.Exec(values...)
		/*
			for idx, val := range values {
				if val != nil {
					fmt.Print(table_fields[idx].Name, " | ", val, " | ")
				}
			}
			fmt.Println()
		*/
		if err != nil {
			panic(err)
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
	err = txn.Commit()
	if err != nil {
		panic(err)
	}
}

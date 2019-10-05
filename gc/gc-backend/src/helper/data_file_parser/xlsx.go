package parser

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"flag"

	pb "git.corp.angel-salon.com/gc/proto"
	"github.com/lib/pq"
	"github.com/tealeg/xlsx"
)

var flagUploadDir = flag.String("uploaddatafile_dir", "./gc-files-upload", "blabla")

type UpdateSQL struct {
	SQL    string
	Values []interface{}
}

func parseXlsxDataFile(fileName string, dbTable *pb.DatabaseTable, tableMeta *pb.TableMetadata, postgresql string) (string, error) {
	filePath := *flagUploadDir + "/" + dbTable.DbName + "/" + fileName
	fmt.Println("start parsing xlsx file " + filePath)
	excel, err := xlsx.OpenFile(filePath)
	if err != nil {
		return "[load error] 打开xlsx文件时出错", err
	}
	fmt.Println("load done")

	stdColumnMap, err := getStdColumnMap(tableMeta)
	if err != nil {
		return "[head error] 标准列信息不合法", err
	}
	headRow := excel.Sheets[0].Rows[0]
	headRowStr := make([]string, len(headRow.Cells))
	for i, title := range headRow.Cells {
		headRowStr[i] = title.Value
	}
	stdToIndexMap, ewb, nsrmc := getStdtoIndexMap(headRowStr, dbTable)
	if ewb != -1 && nsrmc == -1 {
		return "[head error] xlsx缺列", fmt.Errorf("xlsx为二维表，但找不到名为NSRMC、纳税人名称、ZLBSCJUUID或SBUUID的列")
	}
	stdColumns := make([]*stdColumn, len(headRow.Cells))
	var lacks []string
	for name, field := range stdColumnMap {
		if _, ok := stdToIndexMap[name]; !ok {
			lacks = append(lacks, name)
		}
		if field.rowId > 0 && ewb == -1 {
			return "[head error] xlsx缺列", fmt.Errorf("标准表为二维表，但在xlsx中缺少名为EWBHXH或EWBLXH的列")
		}
		index := stdToIndexMap[name]
		stdColumns[index] = field
	}
	if len(lacks) > 0 {
		return "[head error] xlsx缺列", fmt.Errorf("标准列 " + strings.Join(lacks, ",") + " 在xlsx中找不到对应的列")
	}
	fieldList, fieldToIndexMap, idColumnName := getStdFieldList(tableMeta)
	fmt.Println("head done")
	fmt.Println("EWB=", ewb, "NSRMC=", nsrmc, "stdToIndexMap=", stdToIndexMap, "stdColumns=", stdColumns)

	db, err := sql.Open("postgres", postgresql)
	if err != nil {
		return "[sql error] 连接数据库失败", err
	}
	defer db.Close()

	_, err = db.Exec("ALTER TABLE " + pq.QuoteIdentifier(dbTable.DbName) + " ADD COLUMN " + pq.QuoteIdentifier("__数据来源__") + " varchar(1024)")
	if err == nil {
		fmt.Println("SQL Table格式升级：插入数据来源列，删除所有旧数据")
		_, err = db.Exec("DELETE FROM " + pq.QuoteIdentifier(dbTable.DbName))
		if err != nil {
			fmt.Println(err)
		}
	}

	_, err = db.Exec("DELETE FROM "+pq.QuoteIdentifier(dbTable.DbName)+" WHERE "+pq.QuoteIdentifier("__数据来源__")+"=$1", fileName)
	if err != nil {
		return "[sql error] 删除旧数据失败", err
	}

	txn, err := db.Begin()
	if err != nil {
		return "[sql error] 构造Tx时出错", err
	}

	var totalInserted, typeFloatFail, typeStringFail, typeEwbFail, typeDateFail int
	var detailFLoatFail, detailStringFail, detailEwbFail, detailDateFail string

	// defer func() {
	// 	if err == nil {
	// 		fmt.Println("Commiting", totalInserted, detailFLoatFail, detailStringFail, detailEwbFail, detailDateFail)
	// 		err = txn.Commit()
	// 		fmt.Println("Insert Commited", err)
	// 	} else {
	// 		fmt.Println("Rollbacking")
	// 		err = txn.Rollback()
	// 		fmt.Println("Rollbacked", err)
	// 	}
	// }()

	fieldList = append(fieldList, "__数据来源__")
	stmt, err := txn.Prepare(pq.CopyIn(dbTable.DbName, fieldList...))
	if err != nil {
		return "[sql error] 构造Stmt时出错", err
	}
	fmt.Println("sql prepare done")

	values := make([]interface{}, len(fieldList))
	for i := range values {
		values[i] = nil
	}
	var uuidMap = make(map[string][]UpdateSQL)
	var updateValueMap = make(map[string]int)
	var updateColumns []string
	var paramsList []string
	var updateValues []interface{}
	for _, sheet := range excel.Sheets {
		if len(sheet.Rows) <= 1 {
			continue
		}
		ID := ""
		for rowIdx, row := range sheet.Rows {
			if rowIdx == 0 {
				continue
			}
			shouldInsert := true
			//二维表唯一项：资料报送采集UUID、申报UUID
			if ewb != -1 && rowIdx < len(sheet.Rows)-1 {
				nextRow := sheet.Rows[rowIdx+1]
				now := row.Cells[nsrmc].String()
				ID = now
				next := nextRow.Cells[nsrmc].String()
				if now == next {
					shouldInsert = false
				}
			}

			ewIndex := 0
			if ewb != -1 {
				ewIndex, err = row.Cells[ewb].Int()
				if err != nil {
					str := row.Cells[ewb].String()
					ewIndex64, err := strconv.ParseInt(str, 10, 64)
					if err != nil {
						typeEwbFail++
						if typeEwbFail == 1 {
							detailEwbFail = fmt.Sprintf("第%d行\"%s\"", rowIdx, str)
						}
					} else {
						ewIndex = int(ewIndex64)
					}
				}
			}

			for i, cell := range row.Cells {
				if i >= len(stdColumns) {
					break
				}
				if stdColumns[i] == nil {
					continue
				}
				var val interface{}
				switch stdColumns[i].tp {

				case pb.Type_DATETIME:
					val = 0
					var num float64
					num, err = cell.Float()
					if err != nil {
						str := cell.String()

						parse_sucess := false
						str = strings.Replace(str, "上午", "AM", -1)
						str = strings.Replace(str, "下午", "PM", -1)
						str = TIME_REMOVE_REG.ReplaceAllLiteralString(str, "")
						for _, layout := range TIME_LAYOUTS {
							t, err := time.Parse(layout, str)
							if err == nil {
								parse_sucess = true
								val = t.Unix()
								break
							}
						}
						if !parse_sucess {
							typeDateFail++
							if typeDateFail == 1 {
								detailDateFail = fmt.Sprintf("第%d行第%d列\"%s\"", rowIdx, i+1, cell.Value)
							}
						}
						break

					}
					val = num

					val = xlsx.TimeFromExcelTime(num, false).Unix()
					if strings.HasSuffix(stdColumns[i].name, "年月") {
						date := int32(num)
						year := date / 100
						month := date % 100
						layout := "2006-01-02T15:04:05.000Z"
						str := fmt.Sprintf("%d-%02d-01T00:00:00.000Z", year, month)
						t, err := time.Parse(layout, str)
						if err != nil {
							typeDateFail++
							if typeDateFail == 1 {
								detailDateFail = fmt.Sprintf("第%d行第%d列\"%v\"", rowIdx, i+1, num)
							}
							break
						}
						val = t.Unix()
					}
				case pb.Type_FLOAT:
					val, err = cell.Float()
					if err != nil {
						typeFloatFail++
						val = 0
						if typeFloatFail == 1 {
							detailFLoatFail = fmt.Sprintf("第%d行第%d列\"%v\"", rowIdx, i+1)
						}
						break
					}
				case pb.Type_STRING:
					val = cell.String()
						typeStringFail++
						val = ""
						if typeStringFail == 1 {
							detailStringFail = fmt.Sprintf("第%d行第%d列", rowIdx, i+1)
						}
						break
				}

				col := stdColumns[i].name
				if stdColumns[i].rowId > 0 {
					if ewIndex == 0 {
						return "[data error] 二维表行序号匹配失败", fmt.Errorf("二维表序号%d不匹配标准列信息的序号%d", ewIndex, stdColumns[i].rowId)
					}
					found := false
					for column, field := range stdColumnMap {
						if field.rowId == ewIndex && stdToIndexMap[column] == i {
							col = column
							found = true
							break
						}
					}
					if !found {
						return "找不到匹配的列", fmt.Errorf("二维表无法匹配，行号%d, 列号%d 二位行序号 %d 标准行序号 %d", rowIdx, i, ewIndex, stdColumns[i].rowId)
					}
				}
				if index, ok := fieldToIndexMap[col]; ok {
					if totalInserted <= 3 {
						fmt.Println("put val: %s from line %d col (%d, %s) to %d", val, rowIdx, i, col, index)
					}
					values[index] = val
					_, has := updateValueMap[fieldList[index]]
					if _, ok := uuidMap[ID]; !has && ok && val != nil {
						updateValues = append(updateValues, val)
						updateColumns = append(updateColumns, pq.QuoteIdentifier(fieldList[index]))
						paramsList = append(paramsList, "$"+strconv.Itoa(len(paramsList)+1))
						updateValueMap[fieldList[index]] = 1
					}

				} else {
					return "[data error] 标准列信息对应出错", fmt.Errorf("找不到名为%s的标准列", col)
				}
			}

			if shouldInsert {
				values[len(values)-1] = fileName
				totalInserted++
				if totalInserted <= 3 {
					fmt.Println(rowIdx, values)
				}
				if _, ok := uuidMap[ID]; ok && len(updateValues) > 0 {
					updateSQL := "UPDATE " + pq.QuoteIdentifier(dbTable.DbName) + " SET (" + strings.Join(updateColumns, ",") + ") = (" + strings.Join(paramsList, ",") + ") WHERE " + pq.QuoteIdentifier(idColumnName) + " = $" + strconv.Itoa(len(paramsList)+1)
					updateValues = append(updateValues, ID)
					uuidMap[ID] = append(uuidMap[ID], UpdateSQL{SQL: updateSQL, Values: updateValues})
					paramsList = []string{}
					updateColumns = []string{}
					updateValues = []interface{}{}
					updateValueMap = make(map[string]int)
					//update  需要列名、params、values
				} else {
					_, err = stmt.Exec(values...)
					if err != nil {
						fmt.Println("err during insertion")
						return "[data error] Stmt执行插入时出错", err
					}
					if ID!=""{
						uuidMap[ID] = []UpdateSQL{}
					}
				}
				for i := range values {
					values[i] = nil
				}
			}
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return "[sql error] Stmt Exec时出错", err
	}

	err = stmt.Close()
	if err != nil {
		return "[sql error] Stmt close时出错", err
	}

	fmt.Println("Commiting", totalInserted, detailFLoatFail, detailStringFail, detailEwbFail, detailDateFail)
	err = txn.Commit()
	if err != nil {
		txn.Rollback()
		return "[sql error] Tx commit 时出错", err
	}
	txn, err = db.Begin()
	for _, sqls := range uuidMap {
		for _, sqlStruct := range sqls {
			_, err = txn.Exec(sqlStruct.SQL, sqlStruct.Values...)
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		txn.Commit()
	} else {
		txn.Rollback()
		return "[sql error] update 时出错", err
	}

	matchingResult := ""
	for col, val := range stdColumnMap {
		matchingResult = matchingResult + col + ": " + strconv.FormatInt(int64(stdToIndexMap[col]), 10) + "  " + strconv.FormatInt(int64(val.rowId), 10) + "\n"
	}

	return matchingResult + fmt.Sprintf("成功插入%d行，每行包括%d列，共%d个数据，其中有%d个转换成实数时失败，发生错误的数据为%s，有%d个转换成字串时失败，发生错误的数据为%s，有%d个二维表序号提取失败，发生错误的数据为%s，有%d个日期提取失败，发生错误的数据为%s，格式有误的数据使用默认值插入，实数为0，字串为空串。。",
		totalInserted, len(fieldList), totalInserted*len(fieldList),
		typeFloatFail, detailFLoatFail, typeStringFail, detailStringFail, typeEwbFail, detailEwbFail, typeDateFail, detailDateFail), nil
}

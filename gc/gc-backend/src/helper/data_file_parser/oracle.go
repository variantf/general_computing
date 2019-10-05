package parser

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "git.corp.angel-salon.com/gc/proto"
	pq "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
)

func LoadOracleDB(targetConnStr string, targetGroup string, targetTable string,
	dbTable *pb.DatabaseTable, tableMeta *pb.TableMetadata, postgresql string) (string, error) {

	sourceName := "OracleDB:" + targetGroup + ":" + targetTable

	oracle_db, err := sql.Open("oci8", targetConnStr)
	defer oracle_db.Close()
	rows, err := oracle_db.Query("select * from " + targetGroup + "." + targetTable + "order by sbuuid")
	if err != nil {
		rows, err = oracle_db.Query("select * from " + targetGroup + "." + targetTable + "order by zlbscjuuid")
		if err != nil {
			rows, err = oracle_db.Query("select * from " + targetGroup + "." + targetTable)
			if err != nil {
				return "查询表错误", err
			}
		}
	}

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	fmt.Println("load done")

	stdColumnMap, err := getStdColumnMap(tableMeta)
	if err != nil {
		return "[head error] 标准列信息不合法", err
	}
	stdToIndexMap, ewb, nsrmc := getStdtoIndexMap(columns, dbTable)
	if ewb != -1 && nsrmc == -1 {
		return "[head error] db缺列", fmt.Errorf("数据为二维表，但找不到名为NSRMC、纳税人名称、ZLBSCJUUID或SBUUID的列")
	}
	stdColumns := make([]*stdColumn, len(columns))
	var lacks []string
	for name, field := range stdColumnMap {
		if _, ok := stdToIndexMap[name]; !ok {
			lacks = append(lacks, name)
		}
		if field.rowId > 0 && ewb == -1 {
			return "[head error] db缺列", fmt.Errorf("标准表为二维表，但在db中缺少名为EWBHXH或EWBLXH的列")
		}
		index := stdToIndexMap[name]
		stdColumns[index] = field
	}
	if len(lacks) > 0 {
		return "[head error] db缺列", fmt.Errorf("标准列 " + strings.Join(lacks, ",") + " 在csv中找不到对应的列")
	}
	fieldList, fieldToIndexMap, _ := getStdFieldList(tableMeta)

	row_data := make([]interface{}, len(columns))
	row_ptrs := make([]interface{}, len(columns))
	for i, col := range stdColumns {
		row_ptrs[i] = &row_data[i]
		if col == nil {
			continue
		}
		switch col.tp {
		case pb.Type_BOOLEAN:
			row_data[i] = false
		case pb.Type_STRING:
			row_data[i] = ""
		case pb.Type_FLOAT:
			row_data[i] = float64(0)
		case pb.Type_DATETIME:
			row_data[i] = time.Now()
		default:
			return "错误", fmt.Errorf("未知的数据类型: %d", col.tp)
		}
	}

	fmt.Println("head done")
	fmt.Println("EWB=", ewb, "NSRMC=", nsrmc, "stdToIndexMap=", stdToIndexMap)

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

	_, err = db.Exec("DELETE FROM "+pq.QuoteIdentifier(dbTable.DbName)+" WHERE "+pq.QuoteIdentifier("__数据来源__")+"=$1", sourceName)
	if err != nil {
		return "[sql error] 删除旧数据失败", err
	}

	txn, err := db.Begin()
	if err != nil {
		return "[sql error] 构造Tx时出错", err
	}

	defer func() {
		if err == nil {
			fmt.Println("Commiting")
			err = txn.Commit()
			fmt.Println("Commited", err)
		} else {
			fmt.Println("Rollbacking")
			fmt.Println("Rollbacked", err)
		}
	}()

	fieldList = append(fieldList, "__数据来源__")
	stmt, err := txn.Prepare(pq.CopyIn(dbTable.DbName, fieldList...))
	if err != nil {
		return "[sql error] 构造Stmt时出错", err
	}
	fmt.Println("sql prepare done")

	var totalInserted, typeFloatFail, typeStringFail, typeEwbFail, typeDateFail int
	var detailFLoatFail, detailStringFail, detailEwbFail, detailDateFail string
	values := make([]interface{}, len(fieldList))
	for i := range values {
		values[i] = nil
	}
	var rowIdx int
	rows.Next()
	for {
		errRead := rows.Scan(row_ptrs...)
		if errRead != nil {
			return fmt.Sprintf("[read error] 读取db时出错，第%d行", rowIdx), errRead
		}
		row := make([]interface{}, len(row_data))
		nextRow := make([]interface{}, len(row_data))

		for i, value := range row_data {
			row[i] = value
		}
		if !rows.Next() {
			break
		}
		errRead = rows.Scan(row_ptrs...)
		for i, value := range row_data {
			nextRow[i] = value
		}

		shouldInsert := true
		if ewb != -1 && errRead == nil {
			now := row[nsrmc].(string)
			next := nextRow[nsrmc].(string)
			if now == next {
				shouldInsert = false
			}
		}

		ewIndex := 0
		if ewb != -1 {
			ewIndex64, err := strconv.ParseInt(row[ewb].(string), 10, 64)
			if err != nil {
				typeEwbFail++
				if typeEwbFail == 1 {
					detailEwbFail = fmt.Sprintf("第%d行\"%s\"", rowIdx, row[ewb])
				}
			} else {
				ewIndex = int(ewIndex64)
			}
		}

		for i, cell := range row {
			if i >= len(stdColumns) {
				break
			}
			if stdColumns[i] == nil {
				continue
			}
			var val interface{}
			switch stdColumns[i].tp {
			case pb.Type_DATETIME:
				if time, ok := cell.(time.Time); ok {
					val = time.Unix()
				} else {
					val = 0
				}
			case pb.Type_FLOAT:
				val = cell
			case pb.Type_STRING:
				if _, ok := cell.(string); ok {
					val = strings.Replace(cell.(string), "\x00", "", -1)
				} else {
					val = cell
				}
			}
			col := stdColumns[i].name
			if stdColumns[i].rowId > 0 {
				if ewIndex == 0 {
					err = fmt.Errorf("在第%d行，二维表序号%d不匹配标准列信息的序号%d", rowIdx, ewIndex, stdColumns[i].rowId)
					return "[data error] 二维行序号匹配失败", err
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
					fmt.Println("ERROR 3")
					err = fmt.Errorf("二维表无法匹配，行号%d, 二位行序号 %d 标准行序号 %d 标准名称 %s", rowIdx, ewIndex, stdColumns[i].rowId, stdColumns[i].name)
					val = nil
				}
			}
			if index, ok := fieldToIndexMap[col]; ok {
				values[index] = val
			} else {
				err = fmt.Errorf("在第%d行，找不到名为%s的标准列", rowIdx, col)
				return "[data error] 标准列信息对应出错", err
			}
		}

		if shouldInsert {
			values[len(values)-1] = sourceName
			totalInserted++
			if totalInserted <= 3 {
				fmt.Println(rowIdx, values)
			}
			_, err = stmt.Exec(values...)
			if err != nil {
				return "[data error] Stmt执行插入时出错", err
			}
			for i := range values {
				values[i] = nil
			}
		}
	}

	_, err = stmt.Exec()
	if err != nil {

		fmt.Println("ERROR 6")
		return "[sql error] Stmt Exec时出错", err
	}
	err = stmt.Close()
	if err != nil {
		fmt.Println("ERROR 7")
		return "[sql error] Stmt close时出错", err
	}

	fmt.Println("OK last line %s", targetTable)

	return fmt.Sprintf("成功插入%d行，每行包括%d列，共%d个数据，其中有%d个转换成实数时失败，发生错误的数据为%s，有%d个转换成字串时失败，发生错误的数据为%s，有%d个二维表序号提取失败，发生错误的数据为%s，有%d个日期提取失败，发生错误的数据为%s，格式有误的数据使用默认值插入，实数为0，字串为空串。",
		totalInserted, len(fieldList), totalInserted*len(fieldList),
		typeFloatFail, detailFLoatFail, typeStringFail, detailStringFail, typeEwbFail, detailEwbFail, typeDateFail, detailDateFail), nil
}

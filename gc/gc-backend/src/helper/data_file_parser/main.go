package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	pb "git.corp.angel-salon.com/gc/proto"
)

type stdColumn struct {
	name  string
	tp    pb.Type
	rowId int
}

var TIME_LAYOUTS = []string{"2006/1/2 15:04:05", "2006/1/2", "2006-1-2", "2006-1-2 PM 03:04:05", "2006-01-02 PM 03:04:05", "2006-01-02", "2-1月 -06", "2-1月-06", "2-1月 -06 3.4.5.000000000 PM", "2-1月-06 3.4.5.000000000 PM", "2006/1/2 PM 3:4:5", "02-1月 -06 03.04.05.000000 PM", "02-1月-06 03.04.05.000000 PM", "02-1月 -06", "02-1�� -06", "02-1月-06"}
var TIME_REMOVE_REG = regexp.MustCompile("\\ 星期(一|二|三|四|五|六|日)")

// ParseDataFile 解析xlsx/csv文件并插入SQL数据库
func ParseDataFile(fileName string, dbTable *pb.DatabaseTable, tableMeta *pb.TableMetadata, psql string) (string, error) {
	if strings.HasSuffix(fileName, ".xlsx") {
		return parseXlsxDataFile(fileName, dbTable, tableMeta, psql)
	} else if strings.HasSuffix(fileName, ".csv") {
		return parseCsvDataFile(fileName, dbTable, tableMeta, psql)
	} else {
		return "", fmt.Errorf("不支持的文件类型")
	}
}

func LoadFromDB(dbType string, targetConnStr string, targetGroup string, targetTable string, dbTable *pb.DatabaseTable, tableMeta *pb.TableMetadata, psql string) (string, error) {
	switch dbType {
	case "Oracle":
		return LoadOracleDB(targetConnStr, targetGroup, targetTable, dbTable, tableMeta, psql)
		break
	}
	return "错误", fmt.Errorf("不支持的数据库类型")
}

// map<列别名,标准列名(简化)>
func getAliasToStdMap(fields []*pb.DatabaseTable_FieldMapping) map[string]string {
	result := make(map[string]string)
	for _, field := range fields {
		for _, alias := range field.Alias {
			result[alias] = field.StdName
		}
	}
	return result
}

// map<标准列名(简化),标准列结构>
func getStdColumnMap(tableMeta *pb.TableMetadata) (map[string]*stdColumn, error) {
	result := make(map[string]*stdColumn)
	for _, field := range tableMeta.Fields {
		if strings.HasSuffix(field.Hint, "]") && strings.Contains(field.Hint, "[") {
			braPos := strings.LastIndex(field.Hint, "[")
			rowId, err := strconv.ParseInt(field.Hint[braPos+1:len(field.Hint)-1], 10, 64)
			if err != nil {
				return nil, err
			}
			result[field.Name] = &stdColumn{name: field.Name, tp: field.Type, rowId: int(rowId)}
		} else {
			result[field.Name] = &stdColumn{name: field.Name, tp: field.Type}
		}
	}
	return result, nil
}

// 标准列名列表、map<标准列名,序号>
func getStdFieldList(tableMeta *pb.TableMetadata) ([]string, map[string]int, string) {
	var list []string
	result := make(map[string]int)
	var ID = ""
	for i, field := range tableMeta.Fields {
		if field.Name == "资料报送采集UUID" || field.Name == "申报UUID" {
			ID = field.Name
		}
		list = append(list, field.Name)
		result[field.Name] = i
	}
	return list, result, ID
}

func getStdtoIndexMap(headRow []string, dbTable *pb.DatabaseTable) (map[string]int, int, int) {
	aliasToStdMap := getAliasToStdMap(dbTable.Fields)
	result := make(map[string]int)
	ewb, nsrmc := -1, -1
	for alias, std_name := range aliasToStdMap {
		if strings.HasSuffix(alias, "]") && strings.Contains(alias, "[") {
			bracketPosition := strings.LastIndex(alias, "[")
			alias = alias[:bracketPosition]
		}
		for idx, val := range headRow {
			val = strings.TrimLeft(val, " ")
			val = strings.TrimRight(val, " ")
			if val == alias {
				result[std_name] = idx
			}
		}
	}
	for i, val := range headRow {
		val = strings.TrimLeft(val, " ")
		val = strings.TrimRight(val, " ")
		if val == "EWBHXH" || val == "EWBLXH" || val == "二维表行序号" || val == "二维表列序号" {
			ewb = i
		}
		if val == "ZLBSCJUUID" || val == "资料报送采集UUID" || val == "SBUUID" || val == "申报UUID" {
			nsrmc = i
		}
	}
	fmt.Println(result)
	return result, ewb, nsrmc
}

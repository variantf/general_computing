package main

import (
	"flag"
	"fmt"
	"strings"

	pb "git.corp.angel-salon.com/gc/proto"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type twoDimensionalTable struct {
	Name                string
	SplitValues         []string
	SecondKeyName       string
	SecondKeyCandidates []string
	IdenticalFields     []string
}

var known2DTables = []twoDimensionalTable{
	{
		Name:          "SB_CWBB_QYKJZZYBQY_LRB",
		SplitValues:   []string{"BQJE", "SQJE_1"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_QYKJZZYBQY_ZCFZB",
		SplitValues:   []string{"QMYE_ZC", "NCYE_ZC", "QMYE_QY", "NCYE_QY"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32", "33", "34"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name: "SB_CWBB_QYKJZZYBQY_SYZQY",
		SplitValues: []string{"BNJKCG", "BNYYGJ", "BNWFPLY", "BNSYZQYHJ", "SNSSZBHGB", "SNZBGJ", "SNJKCG", "BNZBGJ",
			"SNSYZQYHJ", "SNWFPLY", "SNYYGJ", "BNSSZBHGB"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_QYKJZZYBQY_XJLLB",
		SplitValues:   []string{"BQJE", "SQJE_1"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32",
			"33", "34", "35"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_QYKJZZ_LRB",
		SplitValues:   []string{"BNLJS", "BYS"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_QYKJZZ_XJLLB",
		SplitValues:   []string{"JE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32",
			"33", "34", "35", "36", "37", "38", "39", "40",
			"41", "42", "43", "44", "45", "46", "47", "48",
			"49", "50", "51", "52", "53", "54", "55", "56"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_QYKJZZ_ZCFZB",
		SplitValues:   []string{"NCS_ZC", "QMS_ZC", "QMS_QY", "NCS_QY"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32",
			"33", "34", "35", "36"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_XQYKJZZ_LRB",
		SplitValues:   []string{"BNLJJE", "BYJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_XQYKJZZ_LRB_NB",
		SplitValues:   []string{"SNJE", "BNLJJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_XQYKJZZ_ZCFZB",
		SplitValues:   []string{"NCYE_QY", "QMYE_QY", "NCYE_ZC", "QMYE_ZC"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_XQYKJZZ_XJLLB",
		SplitValues:   []string{"BNLJJE", "BYJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_XQYKJZZ_XJLLB_NB",
		SplitValues:   []string{"BNLJJE", "SNJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name: "SB_SDS_JMCZ_14ND_ZCZJTXQKB",
		SplitValues: []string{"ZZJEZCZZ", "ZZJEBNZJ", "ZZJEZCJS", "SSJEZCJS", "SSJESSGDZJE",
			"SSJEBNJSZJE", "SSJE2014ZJE", "SSJELJZJE", "NSTZJE", "NSTZYY"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27"},
		IdenticalFields: []string{"SBUUID", "PZXH", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_SYYH_XJLLB",
		SplitValues:   []string{"BQJE", "SQJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
	{
		Name:          "SB_CWBB_ZQGS_XJLLB",
		SplitValues:   []string{"BQJE", "SQJE"},
		SecondKeyName: "EWBHXH",
		SecondKeyCandidates: []string{"1", "2", "3", "4", "5", "6", "7", "8",
			"9", "10", "11", "12", "13", "14", "15", "16",
			"17", "18", "19", "20", "21", "22", "23", "24",
			"25", "26", "27", "28", "29", "30", "31", "32",
			"33", "34", "35", "36"},
		IdenticalFields: []string{"ZLBSCJUUID", "SJGSDQ", "XGR_DM", "XGRQ", "LRRQ", "LRR_DM", "SJTB_SJ"},
	},
}

var (
	twoDimensionalTableMap = map[string]twoDimensionalTable{}
	tableMetaMap           = map[string]*pb.TableMetadata{}
	flagStdname            = flag.String("std", "std.xlsx", "standard defination")
	flagGCBackend          = flag.String("backend", "localhost:12100", "backend service")
	flagUpdateTableMeta    = flag.String("update_meta", "false", "whether to update table metadata")
	flagPath               = flag.String("meta_path", "风险控制", "path that meta stored")
	flagDatabasePrefix     = flag.String("database_prefix", "国税", "databse prefix header")
)

func main() {
	flag.Parse()

	for _, table := range known2DTables {
		twoDimensionalTableMap[table.Name] = table
	}

	var metas []*pb.TableMetadata
	if *flagStdname != "" {
		metas = createTablesFromFile(*flagStdname)
		fmt.Println(len(metas), "table metadata created.")
	}

	conn, err := grpc.Dial(*flagGCBackend, grpc.WithInsecure())
	fmt.Print("connected", conn, err)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewDataManagerClient(conn)
	for _, meta := range metas {
		tableMetaMap[meta.Name] = meta
		if *flagUpdateTableMeta == "true" {
			updateTableMetadata(client, meta)
		}
	}
}

func updateTableMetadata(client pb.DataManagerClient, table *pb.TableMetadata) {
	req := pb.UpsertTableMetadataRequest{
		Table:      table,
		RemoveData: true,
	}
	_, err := client.UpsertTableMetadata(context.Background(), &req)
	if err != nil {
		panic(err)
	}

	fieldMappings := make([]*pb.DatabaseTable_FieldMapping, len(table.Fields))
	for i, field := range table.Fields {
		fieldMappings[i] = &pb.DatabaseTable_FieldMapping{
			StdName: field.Name,
			Alias:   field.Name + "," + field.Hint,
		}
	}
	_, err = client.UpsertDatabaseTable(context.Background(), &pb.DatabaseTable{
		MetaName: table.Name,
		MetaPath: *flagPath,
		DbName:   *flagDatabasePrefix + "_" + table.Name,
		Fields:   fieldMappings,
	})

	fmt.Println("finished", table.Name)
}

func createTablesFromFile(name string) (metas []*pb.TableMetadata) {
	metas = []*pb.TableMetadata{}
	excel, err := xlsx.OpenFile(name)
	if err != nil {
		panic(err)
	}

	lastTableName := ""
	var tableMetadata *pb.TableMetadata
	definationSht := excel.Sheet["2.业务表字段清册"]
	for _, row := range definationSht.Rows[1:] {
		table, err := row.Cells[2].String()
		if err != nil {
			panic(err)
		}
		tableHint, err := row.Cells[1].String()
		if err != nil {
			panic(err)
		}
		column, err := row.Cells[4].String()
		if err != nil {
			panic(err)
		}
		columnHint, err := row.Cells[3].String()
		if err != nil {
			panic(err)
		}
		datatype, err := row.Cells[5].String()
		if err != nil {
			panic(err)
		}
		dtype := pb.Type_BOOLEAN
		if strings.Contains(datatype, "CHAR") {
			dtype = pb.Type_STRING
		} else if strings.Contains(datatype, "NUMBER") || strings.Contains(datatype, "TIME") || strings.Contains(datatype, "DATE") {
			dtype = pb.Type_FLOAT
		} else if strings.Contains(datatype, "CLOB") || strings.Contains(datatype, "BLOB") {
			// skip
		} else {
			fmt.Println("unsupported datatype", datatype)
			continue
		}

		if lastTableName != table {
			if tableMetadata != nil {
				is2d := false
				for _, field := range tableMetadata.Fields {
					if strings.EqualFold(field.GetHint(), "ewbhxh") {
						is2d = true
					}
				}
				if !is2d {
					metas = append(metas, tableMetadata)
				}
			}
			tableMetadata = &pb.TableMetadata{
				Name:   table,
				Path:   *flagPath,
				Fields: []*pb.TableMetadata_Field{},
				Hint:   tableHint,
			}
		}
		_, is2d := twoDimensionalTableMap[table]
		if is2d {
			// for _, field := range tableModel.IdenticalFields {
			// 	if column == field {
			// 		tableMetadata.Fields = append(tableMetadata.Fields, &pb.TableMetadata_Field{
			// 			Name: column,
			// 			Type: dtype,
			// 			Hint: columnHint,
			// 		})
			// 		break
			// 	}
			// }
			// for _, field := range tableModel.SplitValues {
			// 	if column == field {
			// 		for _, suffix := range tableModel.SecondKeyCandidates {
			// 			tableMetadata.Fields = append(tableMetadata.Fields, &pb.TableMetadata_Field{
			// 				Name: column + "[" + suffix + "]",
			// 				Type: dtype,
			// 				Hint: columnHint + "[" + suffix + "]",
			// 			})
			// 		}
			// 	}
			// }
		} else {
			tableMetadata.Fields = append(tableMetadata.Fields, &pb.TableMetadata_Field{
				Name: column,
				Type: dtype,
				Hint: columnHint,
			})
		}
		lastTableName = table
	}
	return
}

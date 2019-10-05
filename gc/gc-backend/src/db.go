package server

import (
	"database/sql"
	"flag"
	"sync"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"gopkg.in/mgo.v2"
)

var (
	session             *mgo.Session
	db                  *sql.DB
	oracle_db           *sql.DB
	mgoOnce             sync.Once
	sqlOnce             sync.Once
	flagPostgreSQL      = flag.String("postgresql", "postgres://postgres:AppStore321@localhost/general_computing?sslmode=disable", "PostgreSQL server config.")
	flagMySQL           = flag.String("mysql", "root:SoulYee2015@tcp(www2.angel-salon.com:63306)/angel", "MySQL server config.")
	flagMongoDB         = flag.String("mongo", "mongodb://localhost/gc", "MongoDB config in format 'IP:port/database'.")
	flagTableMetaColle  = flag.String("tablemeta", "tables", "MongoDB Table-Metadata collection name.")
	flagTableDbColle    = flag.String("tabledb", "tabledb", "MongoDB Table-Database collection name.")
	flagTaskColle       = flag.String("tasks", "tasks", "MongoDB Task collection name.")
	flagPipelineColle   = flag.String("pipelines", "pipelines", "MongoDB Pipeline collection name.")
	flagFileDataColle   = flag.String("datafiles", "datafiles", "MongoDB Data-File collection name.")
	flagDataFileDir     = flag.String("datafile_dir", "./gc-files-upload", "path to the directory which contains xlsx/csv data files.")
	flagOracleMsgColle  = flag.String("oraclemsg", "oraclemsg", "MongoDB OracleMsg collection name.")
	flagRunTimeColle    = flag.String("runtime", "runtime", "MongoDB record run time.")
	flagTestErrColle    = flag.String("TestErr", "testerr", "MongoDB save err.")
	flagTestResultColle = flag.String("TestResult", "testresult", "MongoDB save TestResult.")
	flagOracle          = flag.String("oracle", "C##GS/GS@192.168.143.240:1521/ORCL", "")
)

func initMongo() {
	var err error
	session, err = mgo.Dial(*flagMongoDB)
	if err != nil {
		panic(err)
	}
}

func Mongo() *mgo.Database {
	mgoOnce.Do(initMongo)
	return session.Clone().DB("")
}

func ExistMongo(collection string, selector interface{}) (bool, error) {
	db := Mongo()
	defer db.Session.Close()
	num, err := db.C(collection).Find(selector).Count()
	return num > 0, err
}

func initSQL() {
	var err error
	db, err = sql.Open("mysql", *flagMySQL)
	if err != nil {
		panic(err)
	}
}

func QuerySQL(query string, args ...interface{}) (*sql.Rows, error) {
	sqlOnce.Do(initSQL)
	fmt.Println(query, len(args))
	return db.Query(query, args...)
}

func ExecSQL(query string, args ...interface{}) error {
	sqlOnce.Do(initSQL)
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(args...)
	return err
}

func QueryOracle(query string) (*sql.Rows, error) {
	var err error
	oracle_db, err = sql.Open("oci8", *flagOracle)
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
	return oracle_db.Query(query)
}
func ExecOracle(query string) error {
	var err error
	oracle_db, err = sql.Open("oci8", *flagOracle)
	if err != nil {
		panic(err)
	}
	_, err = oracle_db.Exec(query)
	return err
}

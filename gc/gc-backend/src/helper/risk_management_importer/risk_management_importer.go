package main

import (
	"database/sql"
	"sync"

	_ "github.com/lib/pq"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	session *mgo.Session
	db      *sql.DB
	mgoOnce sync.Once
	sqlOnce sync.Once
)

func initMongo() {
	var err error
	session, err = mgo.Dial("localhost:27017/gs_risk_management")
	if err != nil {
		panic(err)
	}
}

func Mongo() *mgo.Database {
	mgoOnce.Do(initMongo)
	return session.Clone().DB("")
}

func initSQL() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:AppStore321@192.168.44.199/general_computing?sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func QuerySQL(query string, args ...interface{}) (*sql.Rows, error) {
	sqlOnce.Do(initSQL)
	return db.Query(query, args...)
}

func main() {
	db := Mongo()
	defer db.Session.Close()
	db.C("companies").UpdateAll(bson.M{}, bson.M{"$set": bson.M{"taxes": 0}})
	rows, err := QuerySQL("SELECT 公司名称, 税款, 税种, 描述, 年份, 月份 " +
		"FROM (select *, ROW_NUMBER() OVER(PARTITION BY \"公司名称\", \"年份\", \"月份\", \"TASK_NAME\") as rk FROM \"RESULT_TAX\") as foo " +
		"WHERE \"公司名称\" is not null and rk = 1 " +
		" and \"年份\" is not null")
	if err != nil {
		panic(err)
	}
	db.C("results").RemoveAll(bson.M{})
	for rows.Next() {
		var name, taxType, desc string
		var tax, year, mongth float64
		err = rows.Scan(&name, &tax, &taxType, &desc, &year, &mongth)
		if err != nil {
			panic(err)
		}
		db.C("companies").Update(bson.M{"name": name}, bson.M{"$inc": bson.M{"taxes": tax, "taxes_cnt": 1}})
		db.C("results").Insert(bson.M{
			"公司名称": name,
			"年份":   year,
			"税款":   tax,
			"描述":   desc,
			"月份":   mongth,
			"税种":   taxType,
		})
	}
}

package main

import (
	"bell-monitor/util"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	var (
		client     = util.GetMgoCli()
		db         *mongo.Database
		collection *mongo.Collection
	)
	//2.选择数据库 my_db
	db = client.Database("bell")

	//选择表 my_collection
	collection = db.Collection("my_collection")
	collection = collection
}

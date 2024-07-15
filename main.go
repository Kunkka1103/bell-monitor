package main

import (
	"bell-monitor/prometh"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDSN = flag.String("mongo-dsn", getEnv("MONGO_DSN", "mongodb://localhost:27017"), "MongoDB DSN")
var pushAddr = flag.String("push-address", getEnv("PUSH_ADDRESS", ""), "Address of the Pushgateway to send metrics")
var interval = flag.Int("interval", getEnvAsInt("INTERVAL", 1), "Interval in minutes to check the delay")

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(*mongoDSN).SetMaxPoolSize(200)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("连接 MongoDB 出错: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("断开 MongoDB 连接出错: %v", err)
		}
	}()

	collection := client.Database("bell").Collection("Tipset")

	ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
	defer ticker.Stop()

	// 处理优雅停机
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				opCtx, opCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer opCancel()

				minTimestamp, err := fetchMinTimestamp(opCtx, collection)
				if err != nil {
					log.Printf("获取 MinTimestamp 出错: %v", err)
					continue
				}

				currentTime := time.Now()
				diff := currentTime.Sub(minTimestamp).Seconds()
				log.Printf("时间差: %v 秒", diff)
				// 假设 prometh.Push 函数存在并能正确发送数据到 Pushgateway。
				prometh.Push(*pushAddr, diff, "main-net")
			case <-stop:
				done <- true
				return
			}
		}
	}()

	<-done
	log.Println("程序已优雅地退出")
}

func fetchMinTimestamp(ctx context.Context, collection *mongo.Collection) (time.Time, error) {
	var result struct {
		MinTimestamp time.Time `bson:"MinTimestamp"`
	}

	// 仅检索按 `_id` 降序排列的最后一个文档，并仅选择 `MinTimestamp` 字段
	opts := options.FindOne().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetProjection(bson.D{{Key: "MinTimestamp", Value: 1}})

	err := collection.FindOne(ctx, bson.D{}, opts).Decode(&result)
	if err != nil {
		return time.Time{}, err
	}
	return result.MinTimestamp, nil
}

func connectWithRetry(ctx context.Context, dsn string, retries int) (*mongo.Client, error) {
	var client *mongo.Client
	var err error
	for i := 0; i < retries; i++ {
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(dsn))
		if err == nil {
			return client, nil
		}
		log.Printf("连接 MongoDB 失败，重试 %d/%d: %v", i+1, retries, err)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, defaultVal int) int {
	if valueStr, exists := os.LookupEnv(name); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}

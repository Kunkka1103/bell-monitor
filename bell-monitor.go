package main

import (
	"bell-monitor/prometh"
	"context"
	"flag"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDSN = flag.String("mongo-dsn", "mongodb://localhost:27017", "MongoDB DSN")
var pushAddr = flag.String("push-address", "", "Address of the Pushgateway to send metrics")
var interval = flag.Int("interval", 1, "Interval in minutes to check the delay")

func main() {
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoDSN))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	collection := client.Database("bell").Collection("Tipset")

	ticker := time.NewTicker(time.Duration(*interval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		minTimestamp, err := fetchMinTimestamp(ctx, collection)
		if err != nil {
			log.Printf("Error fetching MinTimestamp: %v", err)
			continue
		}

		currentTime := time.Now()
		diff := currentTime.Sub(minTimestamp).Seconds()
		log.Printf("Time difference: %v seconds", diff)
		// Assuming prometh.Push function exists and properly sends the data to Pushgateway.
		prometh.Push(*pushAddr, diff, "main-net")
	}
}

func fetchMinTimestamp(ctx context.Context, collection *mongo.Collection) (time.Time, error) {
	var result struct {
		MinTimestamp time.Time `bson:"MinTimestamp"`
	}

	// Only retrieve the last document sorted by `_id` in descending order and select only the `MinTimestamp` field
	opts := options.FindOne().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetProjection(bson.D{{Key: "MinTimestamp", Value: 1}})  // Correctly setting the projection here

	err := collection.FindOne(ctx, bson.D{}, opts).Decode(&result)
	if err != nil {
		return time.Time{}, err
	}
	return result.MinTimestamp, nil
}

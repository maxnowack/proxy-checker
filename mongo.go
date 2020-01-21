package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func getMongo() *mongo.Collection {
	if client == nil {
		url := os.Getenv("MONGO_URL")
		clientOptions := options.Client().ApplyURI(url)
		clientOptions.SetDirect(true)

		cli, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		err = cli.Ping(context.TODO(), nil)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		client = cli
	}

	return client.Database("proxychecker").Collection("proxies")
}

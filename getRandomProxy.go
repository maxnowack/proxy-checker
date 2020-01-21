package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func findProxies() []ProxyDoc {
	proxies := getMongo()

	leastChecked := time.Now().Add(time.Hour * -1)
	query := bson.M{
		"success": true,
		"checkedAt": bson.M{
			"$gte": primitive.NewDateTimeFromTime(leastChecked),
		},
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"speedDown", -1}})
	findOptions.SetProjection(bson.D{{"url", 1}})
	findOptions.SetLimit(100)

	cur, err := proxies.Find(context.TODO(), query, findOptions)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var docs []ProxyDoc
	for cur.Next(context.TODO()) {
		var proxy ProxyDoc
		err := cur.Decode(&proxy)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		docs = append(docs, proxy)
	}
	cur.Close(context.TODO())
	return docs
}

func getRandomProxy() string {
	// find({ working: true, checkedAt: { $gte: (now - 1h) } }, { sort: { quality: -1 }, limit: 100 })
	proxies := findProxies()
	// proxy := gaussianRandomElement(proxies, 2, 0.001).Interface().(ProxyDoc)
	proxy := randomElement(proxies).Interface().(ProxyDoc)
	return proxy.Url
}

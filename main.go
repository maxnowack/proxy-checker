package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProxyDoc struct {
	Url string `bson:"url" json:"url"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Just using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go startProxyChecking()
	log.Println("started proxy checking")

	log.Println("http server listening on port " + port)
	http.HandleFunc("/", proxyServer)
	http.ListenAndServe(":"+port, nil)
}

func processProxy() {
	proxies := getMongo()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"checkedAt", 1}})
	findOptions.SetProjection(bson.D{{"url", 1}})
	findOptions.SetSkip(int64(randomBetween(0, 100)))
	findOptions.SetLimit(1)

	cur, err := proxies.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for cur.Next(context.TODO()) {
		var proxy ProxyDoc
		err := cur.Decode(&proxy)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		cur.Close(context.TODO())

		if proxy.Url == "" {
			log.Fatal("Empty proxy url!")
			os.Exit(1)
		}

		info := checkProxy(proxy.Url)
		var errString string
		if info.err != nil {
			errString = info.err.Error()
		}
		_, err = proxies.UpdateOne(context.TODO(), bson.M{
			"url": proxy.Url,
		}, bson.M{
			"$set": bson.M{
				"success":          info.success,
				"error":            errString,
				"speedUp":          info.speedUp,
				"speedDown":        info.speedDown,
				"dnsLookup":        info.dnsLookup,
				"tcpConnection":    info.tcpConnection,
				"tlsHandshake":     info.tlsHandshake,
				"serverProcessing": info.serverProcessing,
				"remoteIP":         info.remoteIP,
				"checkedAt":        primitive.NewDateTimeFromTime(time.Now()),
			},
		})

		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		return
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("No proxies available")
	os.Exit(0)
}

func startProxyChecking() {
	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY"))
	if err != nil {
		log.Fatal("Cannot parse CONCURRENCY", os.Getenv("CONCURRENCY"))
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func(worker int) {
			for {
				defer wg.Done()
				processProxy()
				wg.Add(1)
			}
		}(worker)
	}

	wg.Wait()
}

func proxyServer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/getRandomProxy" {
		fmt.Fprintf(w, getRandomProxy())
		return
	}

	if r.URL.Path == "/checkProxy" {
		response, err := getProxyCheckResponse(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, response)
		return
	}

	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

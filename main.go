package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProxyDoc struct {
	_id primitive.ObjectID
	url string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", proxyServer)
	go http.ListenAndServe(":"+port, nil)
	log.Println("http server listening on port " + port)

	log.Println("started proxy checking")
	startProxyChecking()
}

func startProxyChecking() {
	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY"))
	if err != nil {
		log.Fatal("Cannot parse CONCURRENCY", os.Getenv("CONCURRENCY"))
		os.Exit(1)
	}

	db := getMongo()
	proxies := db.Collection("proxies")

	sem := make(chan bool, concurrency)
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	for {
		sem <- true
		go func() {
			defer func() { <-sem }()

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			opts := options.FindOne()
			opts.SetSort(bson.D{{"checkedAt", 1}})
			var proxy ProxyDoc
			proxies.FindOne(ctx, bson.M{}, opts).Decode(&proxy)

			info := checkProxy(proxy.url)

			_, err := proxies.UpdateOne(ctx, bson.M{
				"_id": proxy._id,
			}, bson.D{
				{"$set", bson.D{
					{"success", info.success},
					{"error", info.err},
					{"speedUp", info.speedUp},
					{"speedDown", info.speedDown},
					{"dnsLookup", info.dnsLookup},
					{"tcpConnection", info.tcpConnection},
					{"tlsHandshake", info.tlsHandshake},
					{"serverProcessing", info.serverProcessing},
					{"contentTransfer", info.contentTransfer},
					{"checkedAt", primitive.NewDateTimeFromTime(time.Now())},
				}},
			})

			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}()
	}
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

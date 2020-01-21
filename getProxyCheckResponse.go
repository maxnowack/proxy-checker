package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ResponseData struct {
	RequestTime float64 `bson:"requestTime" json:"requestTime"`
	CurrentTime float64 `bson:"currentTime" json:"currentTime"`
	RemoteIP    string  `bson:"remoteIp" json:"remoteIp"`
	Fill        string  `bson:"fill" json:"fill"`
}

func getTimestampMs() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / 1000000.0
}

var responseLength = 1024 * 1024 // 1MB

func getProxyCheckResponse(r *http.Request) (string, error) {
	defer r.Body.Close()
	jsonStr, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var data RequestData
	json.Unmarshal([]byte(jsonStr), &data)

	currentTime := getTimestampMs()
	remoteIp := r.Header.Get("x-forwarded-for")

	log.Println("request from ", remoteIp)

	respData := ResponseData{
		currentTime - data.CurrentTime,
		currentTime,
		remoteIp,
		"",
	}
	jsonData, err := json.Marshal(respData)
	if err != nil {
		return "", err
	}

	respData = ResponseData{
		currentTime - data.CurrentTime,
		currentTime,
		remoteIp,
		buildFillString(responseLength - len(jsonData)),
	}
	jsonData, err = json.Marshal(respData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

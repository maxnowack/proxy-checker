package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type ResponseData struct {
	requestTime uint64
	currentTime uint64
	fill        string
}

func getTimestampMs() uint64 {
	now := time.Now()
	return uint64(now.UnixNano() / 1000000)
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

	respData := ResponseData{
		currentTime - data.currentTime,
		currentTime,
		"",
	}
	jsonData, err := json.Marshal(respData)
	if err != nil {
		return "", err
	}

	respData = ResponseData{
		currentTime - data.currentTime,
		currentTime,
		buildFillString(responseLength - len(jsonData)),
	}
	jsonData, err = json.Marshal(respData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

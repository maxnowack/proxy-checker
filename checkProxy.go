package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/tcnksm/go-httpstat"
	"h12.io/socks"
)

type RequestData struct {
	currentTime uint64
	fill        string
}

type returnValue struct {
	success          bool
	err              error
	blockedSites     []string
	speedUp          float64
	speedDown        float64
	dnsLookup        float64
	tcpConnection    float64
	tlsHandshake     float64
	serverProcessing float64
	contentTransfer  float64
}

type speedInfo struct {
	speedUp          float64
	speedDown        float64
	dnsLookup        float64
	tcpConnection    float64
	tlsHandshake     float64
	serverProcessing float64
	contentTransfer  float64
}

var requestLength = 1024 * 1024 // 1MB

func getSocksClient(url string) *http.Client {
	dialSocksProxy := socks.Dial(url)
	tr := &http.Transport{Dial: dialSocksProxy}
	return &http.Client{Transport: tr}
}

func doTestRequest(httpClient *http.Client) (speedInfo, error) {
	currentTime := getTimestampMs()
	var speed speedInfo

	reqData := RequestData{
		currentTime,
		"",
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return speed, err
	}

	reqData = RequestData{
		currentTime,
		buildFillString(requestLength - len(jsonData)),
	}
	jsonData, err = json.Marshal(reqData)
	if err != nil {
		return speed, err
	}

	req, err := http.NewRequest("GET", os.Getenv("ROOT_URL")+"/checkProxy", nil)
	if err != nil {
		return speed, err
	}

	var result httpstat.Result
	ctx := httpstat.WithHTTPStat(req.Context(), &result)
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return speed, err
	}

	defer resp.Body.Close()
	jsonStr, err := ioutil.ReadAll(resp.Body)

	var data ResponseData
	json.Unmarshal([]byte(jsonStr), &data)

	speed = speedInfo{
		speedUp:          float64(1000 / data.requestTime),
		speedDown:        float64(1000 / (getTimestampMs() - data.currentTime)),
		dnsLookup:        float64(result.DNSLookup / time.Millisecond),
		tcpConnection:    float64(result.TCPConnection / time.Millisecond),
		tlsHandshake:     float64(result.TLSHandshake / time.Millisecond),
		serverProcessing: float64(result.ServerProcessing / time.Millisecond),
		contentTransfer:  float64(result.ContentTransfer(time.Now()) / time.Millisecond),
	}

	return speed, nil
}

func checkProxy(url string) returnValue {
	httpClient := getSocksClient(url)
	var retVal returnValue

	info, err := doTestRequest(httpClient)
	if err != nil {
		retVal.success = false
		retVal.err = err
		return retVal
	}

	retVal.success = true
	retVal.speedUp = info.speedUp
	retVal.speedDown = info.speedDown
	retVal.dnsLookup = info.dnsLookup
	retVal.tcpConnection = info.tcpConnection
	retVal.tlsHandshake = info.tlsHandshake
	retVal.serverProcessing = info.serverProcessing
	retVal.contentTransfer = info.contentTransfer

	// check sites (parallel (sync.WaitGroup))

	return retVal
}

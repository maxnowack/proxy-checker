package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/tcnksm/go-httpstat"
	"h12.io/socks"
)

type RequestData struct {
	CurrentTime float64 `bson:"currentTime" json:"currentTime"`
	Fill        string  `bson:"fill" json:"fill"`
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
	remoteIP         string
}

type speedInfo struct {
	speedUp          float64
	speedDown        float64
	dnsLookup        float64
	tcpConnection    float64
	tlsHandshake     float64
	serverProcessing float64
	remoteIP         string
}

var requestLength = 1024 * 1024 // 1MB

func getSocksClient(url string) *http.Client {
	dialSocksProxy := socks.Dial(url + "/?timeout=30s")
	tr := &http.Transport{Dial: dialSocksProxy}
	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
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

	req, err := http.NewRequest("POST", os.Getenv("ROOT_URL")+"/checkProxy", nil)
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
	json.Unmarshal(jsonStr, &data)

	if data.RequestTime == 0 {
		return speed, errors.New("invalid response")
	}

	speed = speedInfo{
		speedUp:          float64(1000.0 / data.RequestTime),
		speedDown:        float64(1000.0 / (getTimestampMs() - data.CurrentTime)),
		dnsLookup:        float64(result.DNSLookup / time.Millisecond),
		tcpConnection:    float64(result.TCPConnection / time.Millisecond),
		tlsHandshake:     float64(result.TLSHandshake / time.Millisecond),
		serverProcessing: float64(result.ServerProcessing / time.Millisecond),
		remoteIP:         data.RemoteIP,
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
	retVal.remoteIP = info.remoteIP

	// check sites (parallel (sync.WaitGroup))

	return retVal
}

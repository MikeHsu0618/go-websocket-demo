package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const BETFAIR_STREAM_URL = "stream-api.betfair.com:443"
const BETFAIR_SESSION_URL = "https://europe-west2-bf-vendor-app-37822.cloudfunctions.net/betfair-sessionToken-request"
const BETFAIR_APP_KEY = "PUwqsW5mxVDdsEoU"

type Authentication struct {
	Op      string `json:"op"`
	AppKey  string `json:"appKey"`
	Session string `json:"session"`
}

type OrderSubscription struct {
	Op                  string      `json:"op"`
	OrderFilter         OrderFilter `json:"orderFilter"`
	SegmentationEnabled bool        `json:"segmentationEnabled"`
}

type OrderFilter struct {
	IncludeOverallPosition        bool     `json:"includeOverallPosition"`
	CustomerStrategyRefs          []string `json:"customerStrategyRefs"`
	PartitionMatchedByStrategyRef bool     `json:"partitionMatchedByStrategyRef"`
}

type MarketSubscription struct {
	Op               string       `json:"op"`
	Id               int          `json:"id"`
	MarketFilter     MarketFilter `json:"marketFilter"`
	MarketDataFilter struct{}     `json:"marketDataFilter"`
}

type MarketFilter struct {
	MarketIds         []float32 `json:"marketIds"`
	BspMarket         bool      `json:"bspMarket"`
	BettingTypes      []string  `json:"bettingTypes"`
	EventTypeIds      []int     `json:"eventTypeIds"`
	EventIds          []int     `json:"eventIds"`
	TurnInPlayEnabled bool      `json:"turnInPlayEnabled"`
	MarketTypes       []string  `json:"marketTypes"`
	CountryCodes      []string  `json:"countryCodes"`
}

func main() {
	getBetFairSession()
	conn, err := tls.Dial("tcp", BETFAIR_STREAM_URL, &tls.Config{})
	if err != nil {
		log.Println("err", err)
		return
	}
	defer conn.Close()

	go sendRequest(conn)

	for {
		buf := make([]byte, 3000)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Client Received: %v\n", string(buf[:n]))
	}
}

func sendRequest(conn *tls.Conn) {
	auth := Authentication{
		Op:      "authentication",
		AppKey:  BETFAIR_APP_KEY,
		Session: getBetFairSession(),
	}
	authBytes, _ := json.Marshal(auth)

	orderSubscription := OrderSubscription{
		Op: "orderSubscription",
		OrderFilter: OrderFilter{
			IncludeOverallPosition:        false,
			CustomerStrategyRefs:          []string{"betstrategy1"},
			PartitionMatchedByStrategyRef: true,
		},
		SegmentationEnabled: true,
	}
	orderSubscriptionBytes, _ := json.Marshal(orderSubscription)

	marketSubscription := MarketSubscription{
		Op: "marketSubscription",
		Id: 2,
		MarketFilter: MarketFilter{
			MarketIds:         []float32{1.120684740},
			BspMarket:         true,
			BettingTypes:      []string{"ODDS"},
			EventTypeIds:      []int{1},
			EventIds:          []int{27540841},
			TurnInPlayEnabled: true,
			MarketTypes:       []string{"MATCH_ODDS"},
			CountryCodes:      []string{"ES"},
		},
		MarketDataFilter: struct{}{},
	}
	marketSubscriptionBytes, _ := json.Marshal(marketSubscription)

	n, err := conn.Write(authBytes)
	if err != nil {
		log.Println(n, err)
		return
	}
	n, err = conn.Write([]byte("\n"))

	n, err = conn.Write(orderSubscriptionBytes)
	if err != nil {
		log.Println(n, err)
		return
	}
	n, err = conn.Write([]byte("\n"))

	n, err = conn.Write(marketSubscriptionBytes)
	if err != nil {
		log.Println(n, err)
		return
	}
	n, err = conn.Write([]byte("\n"))
}

type Session struct {
	SessionToken string
}

func getBetFairSession() string {
	resp, err := http.Get(BETFAIR_SESSION_URL)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		log.Fatalln(err)
	}

	return session.SessionToken
}

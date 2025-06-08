package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	httpTimeout   = 3
	LedKindWled   = "wled"
	LedKindShelly = "shelly"
	LedKindWiz    = "wiz"
)

var LedKinds = []string{LedKindWled, LedKindShelly}

type Led struct {
	Id      int    `mapstructure:"id"`
	Kind    string `mapstructure:"kind"`
	Address string `mapstructure:"address"`
}

type LedClient struct {
	Leds       map[int]Led
	ledId      int
	address    string
	kind       string
	uri        string
	httpMethod string
	payload    *bytes.Reader
}

func (lc LedClient) String() string {
	return fmt.Sprintf("Leds %#v", lc.Leds)
}

func (lc *LedClient) Id(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
		return errors.New("invalid led Id")
	}
	currLed, ok := lc.Leds[idInt]
	if ok {
		lc.ledId = currLed.Id
		lc.address = currLed.Address
		lc.payload = bytes.NewReader([]byte(``))
	} else {
		return errors.New("unknonw led id")
	}

	switch currLed.Kind {
	case LedKindWled:
		lc.kind = LedKindWled
	case LedKindShelly:
		lc.kind = LedKindShelly
	// TODO:
	// case LedKindWiz:
	// 	lc.kind = LedKindWiz
	default:
		return errors.New("unknown led kind")
	}

	return nil
}

func (lc *LedClient) On() {
	switch lc.kind {
	case LedKindWled:
		lc.uri = "/json/state"
		lc.payload = bytes.NewReader([]byte(`{"on":true}`))
		lc.httpMethod = http.MethodPost
	case LedKindShelly:
		lc.uri = "/light/0/?turn=on"
		lc.httpMethod = http.MethodGet
	}
}

func (lc *LedClient) Off() {
	switch lc.kind {
	case LedKindWled:
		lc.uri = "/json/state"
		lc.payload = bytes.NewReader([]byte(`{"on":false}`))
		lc.httpMethod = http.MethodPost
	case LedKindShelly:
		lc.uri = "/light/0/?turn=off"
		lc.httpMethod = http.MethodGet
	}
}

func (lc *LedClient) Toggle() {
	switch lc.kind {
	case LedKindWled:
		lc.uri = "/json/state"
		lc.payload = bytes.NewReader([]byte(`{"on":"t"}`))
		lc.httpMethod = http.MethodPost
	case LedKindShelly:
		lc.uri = "/light/0/?turn=toggle"
		lc.httpMethod = http.MethodGet
	}
}

func (lc *LedClient) Do() (*http.Response, error) {
	client := &http.Client{
		Timeout: httpTimeout * time.Second,
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		lc.httpMethod,
		"http://"+lc.address+lc.uri,
		lc.payload)
	if err != nil {
		log.Print(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "X-Led-Lighter")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

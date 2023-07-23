package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	TickRate   time.Duration
	HTTPClient *http.Client
	Name       string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SendTemperatureOverHTTP(url string) {
	out := c.createNewTemperature()

	o, err := json.Marshal(out)

	if err != nil {
		log.Fatalln("PrepareTemperature", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(o)))

	if err != nil {
		log.Println("SendTemperatureOverHTTP (New)", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// send request
	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		log.Println("SendTemperatureOverHTTP (Do)", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error sending temperature", resp.StatusCode)
	}
}

func (c *Client) SetTickRate(d time.Duration) {
	c.TickRate = d
}

func (c *Client) SetName(name string) {
	c.Name = name
}

func (c *Client) createNewTemperature() *Temperature {
	return NewTemperature(c.Name, AverageTemperature(ParseTemperatureOutput(CheckTemp())))
}

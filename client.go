package main

import (
	"log"
	"net/http"
	"strings"
)

type Client struct {
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{},
	}
}

func (c *Client) SendTemperatureOverHTTP(t []byte) {
	req, err := http.NewRequest(http.MethodPost, *url, strings.NewReader(string(t)))
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

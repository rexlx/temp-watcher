package main

import (
	"flag"
	"time"
)

var (
	url     = flag.String("url", "http://localhost:8080/temperature", "URL to send temperature to")
	server  = flag.Bool("server", false, "run as server")
	port    = flag.String("port", "8080", "port to run on")
	name    = flag.String("name", "", "name of the app")
	poll    = flag.Duration("poll", 5*time.Second, "how often to poll for temperature")
	metrics = flag.Duration("metrics", 600*time.Second, "how often to print metrics")
)

type Runtime struct {
	URL      string
	Duration time.Duration
	Server   *Server
	Client   *Client
	Hostname string
}

func NewRuntime(name, url string) *Runtime {
	return &Runtime{
		URL:      url,
		Hostname: GetHostnameOrDie(name),
	}
}

func main() {
	flag.Parse()
	run := NewRuntime(*name, *url)
	if *server {
		run.Server = NewServer(*port)
		run.Server.SetTickRate(*metrics)
		run.Server.SetName(run.Hostname)
		run.Server.Forever()
	} else {
		run.Client = NewClient()
		run.Client.SetTickRate(*poll)
		run.Client.SetName(run.Hostname)
		for range time.Tick(run.Client.TickRate) {
			run.Client.SendTemperatureOverHTTP(run.URL)
		}
	}
}

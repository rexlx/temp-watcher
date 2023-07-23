package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
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

type Temperature struct {
	Name string    `json:"name"`
	Temp float64   `json:"temp"`
	Time time.Time `json:"time"`
}

type Runtime struct {
	URL      string
	Duration time.Duration
	Server   *Server
	Client   *Client
	Hostname string
}

type Server struct {
	Mu           sync.Mutex
	TempsPosted  int
	Port         string
	Temperatures []Temperature
}

type Client struct {
	HTTPClient *http.Client
}

func main() {
	flag.Parse()
	run := NewRuntime(*name)
	if *server {
		run.Server = NewServer(*url, *port)
		run.Server.Forever()
	} else {
		client := NewClient()
		for range time.Tick(*poll) {
			client.SendTemperatureOverHTTP(PrepareTemperature())
		}
	}
}

func NewServer(url string, port string) *Server {
	return &Server{
		Port:         port,
		Temperatures: []Temperature{},
	}
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{},
	}
}

func (s *Server) Forever() {
	log.Println("Running as server", s.Port)
	go func() {
		for range time.Tick(*metrics) {
			log.Println("Temperatures posted:", s.TempsPosted)
			log.Println("Average temperature:", s.GetAverageTemperature())
		}
	}()
	http.HandleFunc("/temperature", s.RecieveTemperatureOverHTTP)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.Port), nil))
}

func (s *Server) GetAverageTemperature() float64 {
	if len(s.Temperatures) == 0 {
		return 0
	}
	var sum float64
	for _, temp := range s.Temperatures {
		sum += temp.Temp
	}
	return sum / float64(len(s.Temperatures))
}

func NewRuntime(name string) *Runtime {
	return &Runtime{
		Hostname: GetHostnameOrDie(name),
	}
}

func CheckTemp() []byte {
	// Run the "sensors -f" command.
	out, err := exec.Command("sensors", "-f").Output()
	if err != nil {
		log.Fatalln("CheckTemp", err)
	}

	// Print the output of the command.
	return out
}

func SpaceFieldsJoin(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

func ParseTemperatureOutput(output []byte) []float64 {
	var out []float64
	var i float64
	// look for lines that start with "Core" followed by some integer
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Core") {
			parts := strings.Split(SpaceFieldsJoin(line), " ")
			if _, err := fmt.Sscanf(parts[2], "+%f°F", &i); err != nil {
				log.Println("ParseTemperatureOutput", err, parts)
			} else {
				out = append(out, i)
			}
		}
	}
	return out
}

func AverageTemperature(temps []float64) float64 {
	var sum float64
	for _, temp := range temps {
		sum += temp
	}
	return sum / float64(len(temps))
}

func GetHostnameOrDie(def string) string {
	if def != "" {
		return def
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("GetHostnameOrDie", err)
	}
	return hostname
}

func NewTemperature(temp float64) *Temperature {
	return &Temperature{
		Name: GetHostnameOrDie(*name),
		Temp: temp,
		Time: time.Now(),
	}
}

func PrepareTemperature() []byte {
	out := NewTemperature(AverageTemperature(ParseTemperatureOutput(CheckTemp())))
	o, e := json.Marshal(out)
	if e != nil {
		log.Fatalln("PrepareTemperature", e)
	}
	return o
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

func (s *Server) AddTemperature(t Temperature) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if len(s.Temperatures) > 10000 {
		SaveTemperatures(s.Temperatures)
		s.Temperatures = nil
	}
	s.Temperatures = append(s.Temperatures, t)
}

func (s *Server) GetTemperatures() []Temperature {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Temperatures
}

func SaveTemperatures(tmps []Temperature) {
	type X struct {
		Time         time.Time     `json:"time"`
		Temperatures []Temperature `json:"temperatures"`
	}
	var output X
	// create timestamp for file
	output.Time = time.Now()
	output.Temperatures = tmps
	// save to file
	f, err := os.OpenFile("temps.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("SaveTemperatures (Open)", err)
	}
	defer f.Close()
	out, err := json.Marshal(output)
	if err != nil {
		log.Fatalln("SaveTemperatures (Marshal)", err)
	}
	_, err = f.Write(out)
	if err != nil {
		log.Fatalln("SaveTemperatures (Write)", err)
	}
}

func (s *Server) RecieveTemperatureOverHTTP(w http.ResponseWriter, r *http.Request) {
	var tmp Temperature
	err := json.NewDecoder(r.Body).Decode(&tmp)
	if err != nil {
		log.Fatalln("RecieveTemperatureOverHTTP", err)
	}
	s.TempsPosted++
	s.AddTemperature(tmp)
	// send response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

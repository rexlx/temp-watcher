package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Server is a struct that holds the state of the server
type Server struct {
	Mu           sync.Mutex
	TempsPosted  int
	Port         string
	TickRate     time.Duration
	Temperatures []Temperature
	Name         string
}

// safely add a temperature to the slice. If the slice is too long, save it to a file and clear it.
func (s *Server) AddTemperature(t Temperature) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if len(s.Temperatures) > 10000 {
		SaveTemperatures(s.Temperatures)
		s.Temperatures = nil
	}
	s.Temperatures = append(s.Temperatures, t)
}

func (s *Server) SetTickRate(d time.Duration) {
	s.TickRate = d
}

func (s *Server) SetName(name string) {
	s.Name = name
}

// GetTemperatures returns the slice of temperatures
func (s *Server) GetTemperatures() []Temperature {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Temperatures
}

// returns an address to a new Server
func NewServer(port string) *Server {
	return &Server{
		Port:         port,
		Temperatures: []Temperature{},
	}
}

// Forever runs the server forever, wonder if we should return an error and return our Listener
func (s *Server) Forever() {
	log.Println("Running as server", s.Port)
	go func() {
		for range time.Tick(s.TickRate) {
			log.Println("Temperatures posted:", s.TempsPosted)
			log.Println("Average temperature:", s.GetAverageTemperature())
		}
	}()
	http.HandleFunc("/temperature", s.RecieveTemperatureOverHTTP)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.Port), nil))
}

func (s *Server) GetAverageTemperature() float64 {
	var sum float64

	s.Mu.Lock()
	defer s.Mu.Unlock()

	if len(s.Temperatures) == 0 {
		return 0
	}

	for _, temp := range s.Temperatures {
		sum += temp.Temp
	}

	return sum / float64(len(s.Temperatures))
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

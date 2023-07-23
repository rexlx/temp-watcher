package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Mu           sync.Mutex
	TempsPosted  int
	Port         string
	Temperatures []Temperature
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

func NewServer(url string, port string) *Server {
	return &Server{
		Port:         port,
		Temperatures: []Temperature{},
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

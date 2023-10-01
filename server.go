package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Server is a struct that holds the state of the server
type Server struct {
	Log          *log.Logger
	Mu           sync.RWMutex
	Storage      *minio.Client
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
	if len(s.Temperatures) > 2999 {
		SaveTemperatures(s.Temperatures)
		s.fPutObj("temperatures", fmt.Sprintf("%v.json", time.Now().Unix()), "temps.json")
		s.Temperatures = nil
		log.Println(len(s.Temperatures))
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
	logger := log.New(os.Stdout, "-> ", log.LstdFlags)
	mc, err := minio.New(uri, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Id, s3Key, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln("NewServer", err)
	}
	return &Server{
		Log:          logger,
		Port:         port,
		Storage:      mc,
		Temperatures: []Temperature{},
	}
}

// Forever runs the server forever, wonder if we should return an error and return our Listener
func (s *Server) Forever() {
	s.Log.Println("Running as server", s.Port)
	go func() {
		for range time.Tick(s.TickRate) {
			s.Log.Println("Temperatures posted:", s.TempsPosted)
			s.Log.Println("Average temperature:", s.GetAverageTemperature())
			s.Log.Println("Max temperature:", s.GetMaxTemperature())
		}
	}()
	http.HandleFunc("/temperature", s.RecieveTemperatureOverHTTP)
	s.Log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", s.Port), nil))
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

func (s *Server) GetMaxTemperature() float64 {
	var max float64

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	for _, temp := range s.Temperatures {
		if temp.Temp > max {
			max = temp.Temp
		}
	}

	return max
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

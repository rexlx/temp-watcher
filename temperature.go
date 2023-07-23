package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Temperature struct {
	Name string    `json:"name"`
	Temp float64   `json:"temp"`
	Time time.Time `json:"time"`
}

func NewTemperature(temp float64) *Temperature {
	return &Temperature{
		Name: GetHostnameOrDie(*name),
		Temp: temp,
		Time: time.Now(),
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

func PrepareTemperature() []byte {
	out := NewTemperature(AverageTemperature(ParseTemperatureOutput(CheckTemp())))
	o, e := json.Marshal(out)
	if e != nil {
		log.Fatalln("PrepareTemperature", e)
	}
	return o
}

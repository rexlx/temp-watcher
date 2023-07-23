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

// return address of new Temperature with our hostname and the temperature
func NewTemperature(name string, temp float64) *Temperature {
	return &Temperature{
		Name: name,
		Temp: temp,
		Time: time.Now(),
	}
}

// CheckTemp runs the "sensors -f" command and returns the output. convert to celsius yourself.
func CheckTemp() []byte {
	// Run the "sensors -f" command.
	out, err := exec.Command("sensors", "-f").Output()
	if err != nil {
		log.Fatalln("CheckTemp", err)
	}

	// Print the output of the command.
	return out
}

// ParseTemperatureOutput takes the output of the "sensors -f" command and returns a slice of floats.
func ParseTemperatureOutput(output []byte) []float64 {
	var out []float64
	var i float64
	for _, line := range strings.Split(string(output), "\n") {
		// look for lines that start with "Core" followed by some integer
		if strings.HasPrefix(line, "Core") {
			// the line contains variable whitespace, so we need to join the fields with a single space
			parts := strings.Split(SpaceFieldsJoin(line), " ")
			// this is how we parse the float from the string
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

// PrepareTemperature returns a slice of bytes that is the JSON representation of the Temperature struct.
func PrepareTemperature(name string) []byte {
	out := NewTemperature(name, AverageTemperature(ParseTemperatureOutput(CheckTemp())))
	o, e := json.Marshal(out)
	if e != nil {
		log.Fatalln("PrepareTemperature", e)
	}
	return o
}

package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// SpaceFieldsJoin takes a string and returns a string with all whitespace replaced with a single space
func SpaceFieldsJoin(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

// GetHostnameOrDie returns the hostname of the machine or the default value if it can't be found.
// or it dies.
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

// SaveTemperatures appends the temperatures to a file: {time: "2019-01-01T00:00:00Z", temperatures: [{name: "hostname", temp: 0.0, time: "2019-01-01T00:00:00Z"}]}
func SaveTemperatures(tmps []Temperature) {
	// quick struct
	type X struct {
		Time         time.Time     `json:"time"`
		Temperatures []Temperature `json:"temperatures"`
	}

	// quicker var
	var output X

	// create timestamp for file and set values
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
	// write to file
	_, err = f.Write(out)
	if err != nil {
		log.Fatalln("SaveTemperatures (Write)", err)
	}
}

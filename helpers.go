package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

func SpaceFieldsJoin(str string) string {
	return strings.Join(strings.Fields(str), " ")
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

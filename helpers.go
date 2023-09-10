package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// "github.com/minio/minio-go/v7"
// "github.com/minio/minio-go/v7/pkg/credentials"

var (
	uri   string = "storage.nullferatu.com:9000"
	s3Id  string = "RMCMjDSBEUnHT0vO"
	s3Key string = "OjtUEjI0KEWLCtvDIYy6DwFQ8Pgo6D3g"
)

type Body struct {
	EventName string
	Key       string
	Records   []struct{}
}

func (s *Server) fPutObj(bucketName string, objectName string, filePath string) {
	start := time.Now()
	n, err := s.Storage.FPutObject(context.Background(), bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %v file in %v", n, time.Since(start))
	os.Remove(filePath)
}

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
	f, err := os.OpenFile("temps.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("SaveTemperatures (Open)", err)
	}

	defer f.Close()

	out, err := json.Marshal(output)
	if err != nil {
		log.Fatalln("SaveTemperatures (Marshal)", err)
	}
	// write to file with newline
	_, err = f.Write(out)
	if err != nil {
		log.Fatalln("SaveTemperatures (Write)", err)
	}
	_, err = f.WriteString("\n")
	if err != nil {
		log.Fatalln("SaveTemperatures (Write newline)", err)
	}
}

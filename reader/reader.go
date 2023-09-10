package reader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var wl int = 10

type TempReading struct {
	Name string  `json:"name"`
	Temp float64 `json:"temp"`
	Time string  `json:"time"`
}

type Application struct {
	Workload []string                    `json:"workload"`
	Mu       *sync.Mutex                 `json:"-"`
	Data     map[time.Time][]TempReading `json:"data"`
	Matches  int                         `json:"matches"`
	Time     string                      `json:"time"`
	Temps    []TempReading               `json:"temperatures"`
}

func (a *Application) ReadJsonFromFile(fname string) {
	var res Application
	// read file with readfile
	out, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	data, err := io.ReadAll(out)
	if err != nil {
		log.Fatal(err)
	}
	// unmarshal json
	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Fatal(err)
	}
	t, err := time.Parse(time.RFC3339, res.Time)
	if err != nil {
		log.Fatal(err)
	}
	a.Mu.Lock()
	// parse string to time.Time
	a.Data[t] = res.Temps
	a.Mu.Unlock()
}

func (a *Application) ReadNFiles(files ...string) {
	var wg sync.WaitGroup
	wg.Add(len(files))

	// loop over workload
	for _, v := range files {
		log.Println(v)
		go func(v string) {
			a.ReadJsonFromFile(v)
			wg.Done()
		}(v)
	}
	wg.Wait()
}

func main() {
	theThing := make(map[time.Time][]TempReading)
	mu := sync.Mutex{}
	var app Application
	app.Data = theThing
	app.Mu = &mu
	// get all files in dir
	files, err := os.ReadDir("temperatures")
	if err != nil {
		log.Fatal(err)
	}
	// store the full path in the workload slice
	for _, v := range files {
		app.Workload = append(app.Workload, fmt.Sprintf("temperatures/%v", v.Name()))
	}
	// chunk over workload
	for i := 0; i < len(app.Workload); i += wl {
		// get the workload chunk
		chunk := app.Workload[i:min(i+wl, len(app.Workload))]
		// read the chunk
		app.ReadNFiles(chunk...)
	}
	// iter over app data and print if temp is high enough
	for k, v := range app.Data {
		for _, v2 := range v {
			if v2.Temp > 135 {
				app.Matches++
				log.Printf("%v > Match found at %v: %v", app.Matches, k, v2)
			}
		}
	}
}

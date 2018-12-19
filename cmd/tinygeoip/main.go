package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/mroth/tinygeoip"
)

func main() {
	var dbPath = flag.String("db", "data/GeoLite2-City.mmdb", "Path for MaxMind database file")
	var originPolicy = flag.String("origin", tinygeoip.DefaultOriginPolicy, `'Access-Control-Allow-Origin' header, empty disables`)
	var port = flag.Int("port", 9000, "Port to listen for connections on")
	var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use, otherwise number of CPUs")
	// var verbose = flag.Bool("verbose", false, "log all requests")

	flag.Parse()
	runtime.GOMAXPROCS(*threads)

	db, err := tinygeoip.NewLookupDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	lh := tinygeoip.NewHTTPHandler(db).SetOriginPolicy(*originPolicy)

	// Logging of connections is disabled
	// Logging of connections is enabled, this may severely impact performance under extremely high utilization

	http.Handle("/", lh)
	log.Println("Listening for connections on port", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/mroth/geominder"
)

func main() {
	var dbPath = flag.String("db", "data/GeoLite2-City.mmdb", "Path of MaxMind GeoIP2/GeoLite2 database")
	var port = flag.Int("port", 9000, "Port to listen for connections on")
	var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use, otherwise number of detected cores")
	//var originPolicy = flag.String("origin", "*", `Value sent in the 'Access-Control-Allow-Origin' header. Set to "" to disable.`)

	flag.Parse()
	runtime.GOMAXPROCS(*threads)

	db, err := geominder.NewLookupDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	lh := geominder.NewHTTPHandler(db)


	http.Handle("/", lh)
	log.Println("Listening for connections on port", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/mroth/geominder"
)

func main() {
	var dbPath = flag.String("db", "data/GeoLite2-City.mmdb", "Path of MaxMind GeoIP2/GeoLite2 database")
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
	if err := http.ListenAndServe(":6666", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

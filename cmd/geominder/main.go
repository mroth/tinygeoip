package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"runtime"

	"github.com/mroth/geominder"
	"github.com/pmylund/go-cache"
)

func main() {
	var dbPath = flag.String("db", "GeoLite2-City.mmdb", "Path of MaxMind GeoIP2/GeoLite2 database")
	var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use, otherwise number of detected cores")
	//var originPolicy = flag.String("origin", "*", `Value sent in the 'Access-Control-Allow-Origin' header. Set to "" to disable.`)

	flag.Parse()
	runtime.GOMAXPROCS(*threads)

	// open database
	db, err := geominder.NewLookupDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create a handler for location lookups
	lh := &geominder.HTTPHandler{
		DB:       db,
		MemCache: cache.New(geominder.DefaultCacheExpiration, geominder.DefaultCacheCleanup),
	}

	log.Println("pretending we got a request")
	ip := net.ParseIP("71.246.111.168")

	loc, err := lh.DB.Lookup(ip)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", loc)

	http.Handle("/", lh)
	if err := http.ListenAndServe(":6666", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

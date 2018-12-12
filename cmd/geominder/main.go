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
	var cacheSize = flag.Uint("cache", geominder.DefaultMaxCacheSize, "Max memory used for cache in MB, 0 disables")
	var originPolicy = flag.String("origin", geominder.DefaultOriginPolicy, `Value for 'Access-Control-Allow-Origin' header, set to "" to disable.`)
	var port = flag.Int("port", 9000, "Port to listen for connections on")
	var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use, otherwise number of detected cores")
	// var verbose = flag.Bool("verbose", false, "log all requests")

	flag.Parse()
	runtime.GOMAXPROCS(*threads)

	db, err := geominder.NewLookupDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	lh := geominder.NewHTTPHandler(db).SetOriginPolicy(*originPolicy)
	if *cacheSize != 0 {
		lh.EnableCacheOfSize(*cacheSize)
	} else {
		lh.DisableCache()
	}

	// Logging of connections is disabled
	// Logging of connections is enabled, this may severely impact performance under extremely high utilization

	http.Handle("/", lh)
	log.Println("Listening for connections on port", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

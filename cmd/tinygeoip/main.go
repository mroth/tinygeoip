// The command line binary application for the tinygeoip API microservice.
package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/mroth/tinygeoip"
)

func main() {
	var dbPath = flag.String("db", "data/GeoLite2-City.mmdb", "Path for MaxMind database file")
	var originPolicy = flag.String("origin", tinygeoip.DefaultOriginPolicy, `'Access-Control-Allow-Origin' header, empty disables`)
	var addr = flag.String("addr", ":9000", "Address to listen for connections on")
	var threads = flag.Int("threads", runtime.NumCPU(), "Number of threads to use, otherwise number of CPUs")
	var verbose = flag.Bool("verbose", false, "Log all requests")

	flag.Parse()
	log.Printf("tinygeoip %v\n", tinygeoip.Version)
	runtime.GOMAXPROCS(*threads)

	db, err := tinygeoip.NewLookupDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Printf(
		"Loaded database %s: %v nodes, built %v\n",
		*dbPath, db.NodeCount(), db.BuildTime(),
	)

	lh := tinygeoip.NewHTTPHandler(db).SetOriginPolicy(*originPolicy)

	if *verbose {
		log.Println("Logging of requests is enabled, this may severely impact performance under high utilization!")
		http.Handle("/", trivialLogger(lh))
	} else {
		http.Handle("/", lh)
	}

	log.Println("Listening for connections on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// trivial logging middleware, you would never actually want to use this in
// production -- this is provided mostly as an example of middleware and for
// debugging purposes
func trivialLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Println(r.RemoteAddr, r.RequestURI)
	})
}

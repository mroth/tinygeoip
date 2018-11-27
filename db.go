package main

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

// LookupDB essentially wraps a `maxminddb.Reader` to query for and retrieve our
// minimal data structure. By querying for less, lookups are faster.
//
// Additionally, this allows us to abstract and separate the DB lookup logic from
// the HTTP handlers.
type LookupDB struct {
	reader *maxminddb.Reader
}

// LookupResult is a minimal set of location information that is queried for and
// returned from our lookups.
//
// DEVS: For possible fields, see https://dev.maxmind.com/geoip/geoip2/web-services/
// TODO: maybe make same as https://github.com/bluesmoon/node-geoip?
type LookupResult struct {
	Country  country  `maxminddb:"country" json:"country"`
	Location location `maxminddb:"location" json:"location"`
}

type country struct {
	ISOCode string `maxminddb:"iso_code" json:"iso_code"`
}

type location struct {
	Latitude  float64 `maxminddb:"latitude" json:"lat"`
	Longitude float64 `maxminddb:"longitude" json:"long"`
	// The approximate accuracy radius, in kilometers, around the
	// latitude and longitude for the geographical entity (country,
	// subdivision, city or postal code) associated with the IP address.
	// We have a 67% confidence that the location of the end-user falls
	// within the area defined by the accuracy radius and the latitude
	// and longitude coordinates.
	Accuracy int `maxminddb:"accuracy_radius" json:"accuracy"`
	// The time zone associated with location, as specified by the IANA
	// Time Zone Database, e.g., “America/New_York”.
	// Timezone string `maxminddb:"time_zone"`
}

// NewLookupDB open a new DB reader.
//
// dbPath must be the path to a valid maxmindDB file with at least city level precision.
func NewLookupDB(dbPath string) (*LookupDB, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &LookupDB{reader: db}, nil
}

// Close closes the underlying database and returns resources to the system.
//
// For current implemetnation, see maxminddb.Reader.Close()
func (l *LookupDB) Close() error {
	return l.reader.Close()
}

// Lookup returns the results for a given IP address, or nil and an error
// if results can not be obtained for some reason.
func (l *LookupDB) Lookup(ip net.IP) (*LookupResult, error) {
	var r LookupResult
	err := l.reader.Lookup(ip, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

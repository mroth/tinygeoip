package tinygeoip

import (
	"strconv"
	"sync"
)

// LookupResult is a minimal set of location information that is queried for and
// returned from our lookups.
type LookupResult struct {
	Country  country  `maxminddb:"country" json:"country"`
	Location location `maxminddb:"location" json:"location"`
}

// DEVS: For possible fields, see https://dev.maxmind.com/geoip/geoip2/web-services/

type country struct {
	// A two-character ISO 3166-1 country code for the country associated with
	// the IP address.
	ISOCode string `maxminddb:"iso_code" json:"iso_code"`
}

type location struct {
	// The approximate latitude of the postal code, city, subdivision or country
	// associated with the IP address.
	Latitude float64 `maxminddb:"latitude" json:"latitude"`
	// The approximate longitude of the postal code, city, subdivision or
	// country associated with the IP address.
	Longitude float64 `maxminddb:"longitude" json:"longitude"`
	// The approximate accuracy radius, in kilometers, around the
	// latitude and longitude for the geographical entity (country,
	// subdivision, city or postal code) associated with the IP address.
	// We have a 67% confidence that the location of the end-user falls
	// within the area defined by the accuracy radius and the latitude
	// and longitude coordinates.
	Accuracy int `maxminddb:"accuracy_radius" json:"accuracy_radius"`
	// The time zone associated with location, as specified by the IANA
	// Time Zone Database, e.g., “America/New_York”.
	// Timezone string `maxminddb:"time_zone"`
}

// FastJSON is an faster alternative to calling json.Marshal for a LookupResult.
//
// Yes, if you look at the code you will likely be horrified. Everything is hard
// coded, so if the format for the LookupResult struct is modified, this will
// have to be edited by hand again. So, why?
//
// json.Marshal is already very fast for a small struct like this. In fact,
// using the popular code generation tools ffjson and easyjson, I was unable to
// get them to be significantly more performant for this data struct, and in
// most cases they were actually slower (lesson: always measure!).
//
// However, by doing this hand-made "artisanal" encode, the end result is about
// 2.5x faster than json/ffjson/easyjson.
//
// Note that implementing MarshalJSON with this does not get the same results,
// since you still lose speed to the initial reflection on interface{} (quite a
// bit more than I would have expected!).
//
// You probably don't need this. I may end up deleting it for maintainability.
// However, if you are really trying to screaming fast speed, and every nanosec
// is critical, it could be useful.
func (lr *LookupResult) FastJSON() []byte {
	b := make([]byte, 0, 128) // 106 is largest test case
	b = append(b, `{"country":{"iso_code":"`...)
	b = append(b, lr.Country.ISOCode...)
	b = append(b, `"},"location":{"latitude":`...)
	b = strconv.AppendFloat(b, lr.Location.Latitude, 'f', -1, 64)
	b = append(b, `,"longitude":`...)
	b = strconv.AppendFloat(b, lr.Location.Longitude, 'f', -1, 64)
	b = append(b, `,"accuracy_radius":`...)
	b = strconv.AppendInt(b, int64(lr.Location.Accuracy), 10)
	b = append(b, `}}`...)
	return b
}

// FasterJSON is like FastJSON but backed by an internal sync.Pool to handle
// allocation of byte slices. If FastJSON scared you, then this will scar you.
//
// This version will return a pointer to the byteslice instead. When you are
// done with the data, you can return the backing array to the pool by sending
// that pointer back to PoolReturn(). BE SURE YOU ARE ACTUALLY DONE WITH IT.
//
// By utilizing that methodology, you can have zero memory allocations overall.
//
// However, this is not recommended unless you *really* need it, since you are
// entering the world of needing to be obsessive about your object lifetimes to
// avoid data race scenarios. Please be very careful and audit and measure your
// code.
//
// TODO: possibly delete this monstrosity once we stabilize API.
func (lr *LookupResult) FasterJSON() *[]byte {
	b := dbFastJSONResultsPool.Get().(*[]byte)
	(*b) = (*b)[:0] // reset slice to enable re-use from pool
	(*b) = append((*b), `{"country":{"iso_code":"`...)
	(*b) = append((*b), lr.Country.ISOCode...)
	(*b) = append((*b), `"},"location":{"latitude":`...)
	(*b) = strconv.AppendFloat((*b), lr.Location.Latitude, 'f', -1, 64)
	(*b) = append((*b), `,"longitude":`...)
	(*b) = strconv.AppendFloat((*b), lr.Location.Longitude, 'f', -1, 64)
	(*b) = append((*b), `,"accuracy_radius":`...)
	(*b) = strconv.AppendInt((*b), int64(lr.Location.Accuracy), 10)
	(*b) = append((*b), `}}`...)
	return b
}

var dbFastJSONResultsPool = sync.Pool{
	New: func() interface{} {
		bs := make([]byte, 0, 128)
		return &bs
	},
}

// PoolReturn returns a byteslice to the backing pool for potential re-use.
//
// Be sure you are really done with it before returning!
func (lr *LookupResult) PoolReturn(b *[]byte) {
	dbFastJSONResultsPool.Put(b)
}

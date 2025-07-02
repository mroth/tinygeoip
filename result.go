package tinygeoip

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

# tinygeoip :dragon:

[![Build Status](https://github.com/mroth/tinygeoip/actions/workflows/build.yml/badge.svg)](https://github.com/mroth/tinygeoip/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mroth/tinygeoip)](https://goreportcard.com/report/github.com/mroth/tinygeoip)
[![Go Reference](https://pkg.go.dev/badge/github.com/mroth/tinygeoip.svg)](https://pkg.go.dev/github.com/mroth/tinygeoip)

A small and fast HTTP based microservice for extremely minimal geoip location
lookups.

It bundles into a ~5MB docker image that can serve over ~250K reqs/sec
(uncached) from my laptop.

## API

The API is intentionally extremely minimal and is designed to return only the
absolutely most frequently needed geographic metadata for IP lookups.

The API has only one endpoint `/`, and you just put the IP address (IPv4 and
IPv6 both accepted) directly in the URI path.

Example:

```json5
// $ curl http://${SERVER_IP}/89.160.20.112
{"country":{"iso_code":"SE"},"location":{"latitude":59.4333,"longitude":18.05,"accuracy_radius":200}}
```

Response reformatted for ease of human reading:

```json5
{
  "country": {
    "iso_code": "SE"        // ISO 3166-1 country code
  },
  "location": {
    "latitude": 59.4333,    // Approximate latitude of IP
    "longitude": 18.05,     // Approximate longitude of IP
    "accuracy_radius": 200  // Accuracy radius, in km, for the location
  }
}
```

## Performance

This package _generally_ favors understandability of code over performance
optimizations. That said, it is written in a way to be fairly highly performant,
and combined with it's minimal nature, it can trivially handle a sustained
150,000 requests/second on my workstation. This actually makes it faster than any
other similar off-the-shelf packages I tested in a quick informal survey.
<small>_(Note: my benchmarking was intentionally not robust, and I'm certainly not
trying to start any microbenchmark wars here.)_</small>

## Running the server

```
Usage of tinygeoip:
  -addr string
        Address to listen for connections on (default ":9000")
  -db string
        Path for MaxMind database file (default "data/GeoLite2-City.mmdb")
  -origin string
        'Access-Control-Allow-Origin' header, empty disables (default "*")
  -verbose
        Log all requests
```

You will need to provide a city-level precision GeoIP2 database file. Free
GeoLite2 versions are available for download from [MaxMindDB].

[MaxMindDB]: https://dev.maxmind.com/geoip/geoip2/geolite2/

## Go library

A Go library is provided for utilizing within native projects. A standard
`http.Handler` interface is utilized for compatibility within standard Go HTTP
middleware setups. Alternately, if you are doing the lookups within your
existing application and not over the HTTP microservice, you can get an average
lookup result in approximately 1.2 microseconds.

For more information, see the [GoDocs].

> [!TIP]
> This is a pre v1.0 package that is exported for convenience but is primarily
> consumed by end-users via the binary releases, therefore there may be breaking
> API changes to the library prior to any v1.0 stable release.

[GoDocs]: https://pkg.go.dev/github.com/mroth/tinygeoip


## Docker Image

A docker image is automatically built from all tagged releases.

To utilize it, be sure to mount your MaxMindDB database as a volume so that the
running container can access it.

_[TODO: provide an example for folks not so familiar with Docker.]_

## Stability

:construction: The current API is considered _unstable_. This is just being
released and I'd like some feedback to make any potential changes before tagging
a `v1.0` which will maintain API stability. 

In other words, comments and feedback wanted!

## Related projects

- [`klauspost/geoip-service`][prj1] is where some of the initial inspiration for
  this was drawn. The primary difference is here we have a significantly more
  minimal API (with integration tests), which removes the need for caching or
  external JSON serialization libraries (things we initially had here but
  removed as the perf tradeoff was not benchmarking as significant versus the
  added complexity). I also wanted the API payload to be much smaller for client
  efficiency.

- [`bluesmoon/node-geoip`][prj2] Seems well received, but uses "somewhere between 512MB and 2GB" of memory, which made it highly unsuitable for my purposes.

[prj1]: https://github.com/klauspost/geoip-service
[prj2]: https://github.com/bluesmoon/node-geoip

## License

Software license available upon request.

All licenses will contain an additional clause similar to:

> "This software is not licensed for usage in any application related to censorship or preventing access to information based on geographic region. Legal action will be pursued against any entity who uses this software to knowingly violate this provision."

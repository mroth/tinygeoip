# tinygeoip WIP

A small and fast HTTP based microservice (as well as native Go library)
for extremely minimal geoip location lookups.

## CLI Tool / Microservice

TODO: Description

TODO: CLI options here

```json5
// $ curl http://${SERVER_IP}/89.160.20.112
{"country":{"iso_code":"SE"},"location":{"latitude":59.4333,"longitude":18.05,"accuracy_radius":200}}
```

Reformatted for ease of human reading:

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

TODO: note perf characteristics here

## Library

A Go library is provided for utilizing within native projects. Additionally, a
standard `http.Handler` interface is provided for bundling the HTTP microservice into
existing http Mux setups.

For more information, see the GoDocs.

## TODO

* Custom struct lookup on db to get a very minimal payload. (√ partially complete, need to finalize API)
* Tests and benchmarks. √
* [PERF] Even more performant caching solution (allegro/bigcache).
* [PERF] Utilize ffjson for json serialization.
* [PERF] investigate fasthttp with sync.Pool.
* Modules && Docker build updated.

https://dev.maxmind.com/geoip/geoipupdate/


## Performance

Comparison:
https://www.npmjs.com/package/geoip-lite

Says 20microsecs for lookup, 6microsec for ipv4.

https://allegro.tech/2016/03/writing-fast-cache-service-in-go.html


## Docker Microservice Container

TODO

## Other projects

- [`klauspost/geoip-service`][prj1] is where some of the initial
  structural inspiration for this was drawn. The primary difference has been in
  having a significantly more minimal API (with tests), and performance tuning.


- [`bluesmoon/node-geoip`][prj2] 
Uses "somewhere between 512MB and 2GB" of memory.

[prj1]: https://github.com/klauspost/geoip-service
[prj2]: https://github.com/bluesmoon/node-geoip

## License

TBD

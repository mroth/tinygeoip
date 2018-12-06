# geominder WIP

Partially based on `https://github.com/klauspost/geoip-service` but with some changes.

* Custom struct lookup on ** db to get a very minimal payload. √
* Some simplification.
* Minor tests and benchmarks.
* Utilize *** for json serialization.
* sync.pool? 
* Docker build updated

Comparison:
https://www.npmjs.com/package/geoip-lite

Says 20microsecs for lookup, 6microsec for ipv4.


Put this in data repo:
This product includes GeoLite2 data created by MaxMind, available from
<a href="http://www.maxmind.com">http://www.maxmind.com</a>.


https://dev.maxmind.com/geoip/geoipupdate/

Names?
- geominder
- geoipfeather
- geoipnano
- nanogeoip
- picogeoip
- geoipico*
- tinygeoip



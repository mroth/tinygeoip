// Package tinygeoip implements a small and fast HTTP based microservice for
// extremely minimal geoip location lookups.
package tinygeoip

// Version of this package, adheres to semantic versioning.
const Version = "0.1.0"

// When releasing a new version, increment the version string above. Version
// numbers should always adhere to semantic versioning.
//
// Change the version number in an atomic commit that does nothing else.
//
//  * the commit message should be "release: v1.2.3"
//  * the commit should then be tagged "v1.2.3"
//
// While there are build scripts that can automate this, I prefer it to
// require human oversight.

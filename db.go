package tinygeoip

import (
	"fmt"
	"net"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

// LookupDB essentially wraps a `maxminddb.Reader` to query for and retrieve our
// minimal data structure. By querying for less, lookups are faster.
type LookupDB struct {
	reader *maxminddb.Reader
}

// NewLookupDB open a new DB reader.
//
// dbPath must be the path to a valid maxmindDB file containing city precision.
func NewLookupDB(dbPath string) (*LookupDB, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &LookupDB{reader: db}, nil
}

// Close closes the underlying database and returns resources to the system.
//
// For current implementation, see maxminddb.Reader.Close().
func (l *LookupDB) Close() error {
	return l.reader.Close()
}

// Lookup returns the results for a given IP address, or an error if results can
// not be obtained for some reason.
func (l *LookupDB) Lookup(ip net.IP) (*LookupResult, error) {
	var r LookupResult
	err := l.lookup(ip, &r)
	return &r, err
}

// FastLookup is a version of Lookup that avoids any memory allocations by
// taking a pointer to a pre-allocated LookupResult to decode into.
//
// You probably don't need to use this unless you are tuning for ludicrous speed
// in combination with a sync.Pool, etc.
func (l *LookupDB) FastLookup(ip net.IP, r *LookupResult) error {
	return l.lookup(ip, r)
}

// oschwald/maxminddb-golang does not generate an error on a failed lookup,
// see: https://github.com/oschwald/maxminddb-golang/issues/41
//
// to work around this, we don't use their Lookup(), but rather check
// LookupOffset() first, and throw our own error if nothing was found, before
// using the offset for a manual Decode().
//
// TODO: this issue has now been resolved in recent versions of maxminddb-golang,
// so we can deprecate this and move to using the new LookupNetwork().
func (l *LookupDB) lookup(ip net.IP, r *LookupResult) error {
	offset, err := l.reader.LookupOffset(ip)
	if err != nil {
		return err
	}
	if offset == maxminddb.NotFound {
		return fmt.Errorf("no match for %v found in database", ip)
	}
	return l.reader.Decode(offset, r)
}

// NodeCount returns the number of nodes from the underlying database metadata.
func (l *LookupDB) NodeCount() uint {
	return l.reader.Metadata.NodeCount
}

// BuildTime returns the timestamp for when the underlying database was built.
func (l *LookupDB) BuildTime() time.Time {
	return time.Unix(int64(l.reader.Metadata.BuildEpoch), 0)
}

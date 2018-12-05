REALDATA_URI  = http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
REALDATA_PATH = data/GeoLite2-City.mmdb
PACKAGES = .# TODO: replace with ./... once in multiple
BINDIR = bin

.PHONY: test bench benchreal realdata clobber

default: $(BINDIR)/geominder

# TODO binary
$(BINDIR):
	mkdir -p $(BINDIR)

$(BINDIR)/%: cmd/%/*.go $(BINDIR)
	go build -o $(BINDIR)/$* $<

# run standard go tests, with race detector active
test:
	go test -cover -race $(PACKAGES)

# run microbenchmarks (using the test database, for testing perf regression)
bench:
	go test -run=Bench -bench=. -benchmem $(PACKAGES) 

# run microbenchmarks, with a full copy of the GeoLite2 city database.
# this gets us more accurate measurements as the database size is larger.
# NOTE: since we don't bundle a production database, the dependencies for this
# task will attempt to download one from the public internet if it isn't already
# stored locally. (~26.5MB download)
benchreal: realdata
	go test -run=Bench -bench=. -benchmem $(PACKAGES) -args -db=$(REALDATA_PATH)

realdata: $(REALDATA_PATH)

$(REALDATA_PATH): 
	# TODO: implement me!
	# curl $(REALDATA_URI) | tar -xzv $(REALDATA_PATH)

clobber:
	rm -r $(BINDIR)
	# rm $(REALDATA_PATH)

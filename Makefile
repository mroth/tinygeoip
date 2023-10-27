# Load .env file if present
ifneq (,$(wildcard ./.env))
    include .env
	export
endif

# Downlad of realdata requires MAXMIND_LICENSE_KEY as environment variable or in .env file
REALDATA_URI  = https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$(MAXMIND_LICENSE_KEY)&suffix=tar.gz
REALDATA_DIR  = ./data
REALDATA_PATH = $(REALDATA_DIR)/GeoLite2-City.mmdb
PACKAGES = .
BINDIR = bin

.PHONY: test bench benchreal realdata check-license image clobber

default: $(BINDIR)/tinygeoip

check-license:
ifndef MAXMIND_LICENSE_KEY
	$(error MAXMIND_LICENSE_KEY is unset or empty)
endif

$(BINDIR):
	mkdir -p $(BINDIR)

$(BINDIR)/%: cmd/%/*.go *.go $(BINDIR)
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
$(REALDATA_PATH): check-license
	mkdir -p $(REALDATA_DIR)
	curl "$(REALDATA_URI)" | tar -xzv --strip-components=1 -C $(REALDATA_DIR)

image:
	docker build -t mrothy/tinygeoip .

clobber:
	rm -rf $(BINDIR)
	rm -rf $(REALDATA_DIR)

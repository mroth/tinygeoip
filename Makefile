# Load .env file if present
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

REALDATA_DIR  = ./data
REALDATA_PATH = $(REALDATA_DIR)/GeoLite2-City.mmdb
PACKAGES = .
BINDIR = bin

.PHONY: test bench benchreal realdata check-license image clobber

default: $(BINDIR)/tinygeoip

# Download of realdata requires account id and license key as environment variable or in .env file
check-license:
ifndef MAXMIND_ACCOUNT_ID
	$(error MAXMIND_ACCOUNT_ID is unset or empty)
endif
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

# Download GeoLite2-City.mmdb using the official geoipupdate Docker image
realdata: $(REALDATA_PATH)
$(REALDATA_PATH):
	$(MAKE) check-license
	mkdir -p $(REALDATA_DIR)
	docker run --rm \
	  -e GEOIPUPDATE_ACCOUNT_ID=$(MAXMIND_ACCOUNT_ID) \
	  -e GEOIPUPDATE_LICENSE_KEY=$(MAXMIND_LICENSE_KEY) \
	  -e GEOIPUPDATE_EDITION_IDS=GeoLite2-City \
	  -v $(REALDATA_DIR):/usr/share/GeoIP \
	  ghcr.io/maxmind/geoipupdate:v7.1

image:
	docker build -t ghcr.io/mroth/tinygeoip .

serve-image: realdata
	docker run --rm -it \
		-v $(REALDATA_DIR):/data \
		-p 9000:9000 \
		ghcr.io/mroth/tinygeoip -db /data/GeoLite2-City.mmdb

clobber:
	rm -rf $(BINDIR)
	rm -rf $(REALDATA_DIR)

VERSION=0.4.31

.PHONY: build
build:
	go build ./...

.PHONY: bundle-latest
bundle-latest:
	./packaging/pfsense/scripts/bundle-latest.sh

VERSION=0.4.13

.PHONY: build
build:
	go build ./...

.PHONY: bundle-latest
bundle-latest:
	./packaging/pfsense/scripts/bundle-latest.sh

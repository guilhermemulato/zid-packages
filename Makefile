VERSION=0.4.65

.PHONY: build
build:
	go build ./...

.PHONY: bundle-latest
bundle-latest:
	./packaging/pfsense/scripts/bundle-latest.sh

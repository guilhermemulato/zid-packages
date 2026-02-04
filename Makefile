VERSION=0.4.45

.PHONY: build
build:
	go build ./...

.PHONY: bundle-latest
bundle-latest:
	./packaging/pfsense/scripts/bundle-latest.sh

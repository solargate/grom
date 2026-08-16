.PHONY: all grom apidoc docs docs-serve web cli android android-apk android-aab android-debug gencerts test test-go test-ui clean

VERSION := $(shell tr -d '[:space:]' < VERSION)
BUILD_NUMBER ?= 1
LDFLAGS := -X github.com/solargate/grom/internal/version.Version=$(VERSION)
FLUTTER_VERSION_FLAGS := --build-name=$(VERSION) --build-number=$(BUILD_NUMBER)

all: grom

grom: apidoc web cli
	cd cmd/grom && go build -ldflags "$(LDFLAGS)" -o grom

cli:
	cd cmd/grom && go build -ldflags "$(LDFLAGS)" -o grom

apidoc:
	cd api && swag init -d v1

docs:
	mkdocs build --strict
	cp site/privacy/index.html site/privacy.html

docs-serve:
	mkdocs serve

web:
	cd ui/grom && flutter build web $(FLUTTER_VERSION_FLAGS)
	rm -rf internal/web/dist
	cp -r ui/grom/build/web internal/web/dist
	touch internal/web/dist/.gitkeep

android: android-apk

android-apk:
	cd ui/grom && flutter build apk --release $(FLUTTER_VERSION_FLAGS)

android-aab:
	cd ui/grom && flutter build appbundle --release $(FLUTTER_VERSION_FLAGS)

android-debug:
	cd ui/grom && flutter build apk --debug $(FLUTTER_VERSION_FLAGS)

gencerts:
	cd cmd/grom && go run . gencerts --ip $(IP) --domain $(DOMAIN)

test: test-go test-ui

test-go:
	go test ./...

test-ui:
	cd ui/grom && flutter test

clean:
	rm -f cmd/grom/grom
	rm -rf site
	cd ui/grom && flutter clean

.PHONY: all grom apidoc docs docs-serve venv catalog web cli android android-apk android-aab android-debug gencerts test test-go test-ui test-scripts clean

VERSION := $(shell tr -d '[:space:]' < VERSION)
BUILD_NUMBER ?= 1
LDFLAGS := -X github.com/solargate/grom/internal/version.Version=$(VERSION)
FLUTTER_VERSION_FLAGS := --build-name=$(VERSION) --build-number=$(BUILD_NUMBER)

PYTHON_VENV := .venv
PYTHON := $(PYTHON_VENV)/bin/python
MKDOCS := $(PYTHON_VENV)/bin/mkdocs
VENV_STAMP := $(PYTHON_VENV)/.installed

all: grom

grom: apidoc web cli
	cd cmd/grom && go build -ldflags "$(LDFLAGS)" -o grom

cli:
	cd cmd/grom && go build -ldflags "$(LDFLAGS)" -o grom

apidoc:
	cd api && swag init -d v1

$(VENV_STAMP): requirements.txt
	python3 -m venv $(PYTHON_VENV)
	$(PYTHON_VENV)/bin/pip install -r requirements.txt
	touch $(VENV_STAMP)

venv: $(VENV_STAMP)

docs: venv
	$(MKDOCS) build --strict

docs-serve: venv
	$(MKDOCS) serve

catalog: venv
	$(PYTHON) scripts/server_catalog.py generate

web: catalog
	cd ui/grom && flutter build web $(FLUTTER_VERSION_FLAGS) --no-web-resources-cdn
	rm -rf internal/web/dist
	cp -r ui/grom/build/web internal/web/dist
	touch internal/web/dist/.gitkeep

android: android-apk

android-apk: catalog
	cd ui/grom && flutter build apk --release $(FLUTTER_VERSION_FLAGS)

android-aab: catalog
	cd ui/grom && flutter build appbundle --release $(FLUTTER_VERSION_FLAGS)

android-debug: catalog
	cd ui/grom && flutter build apk --debug $(FLUTTER_VERSION_FLAGS)

gencerts:
	cd cmd/grom && go run . gencerts --ip $(IP) --domain $(DOMAIN)

test: test-go test-ui test-scripts

test-go:
	go test ./...

test-scripts: venv
	scripts/changelog_notes_test.sh
	$(PYTHON) scripts/server_catalog_test.py

test-ui:
	cd ui/grom && flutter test

clean:
	rm -f cmd/grom/grom
	rm -rf site
	cd ui/grom && flutter clean

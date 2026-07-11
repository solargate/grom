.PHONY: all grom doc web cli android android-apk android-aab android-debug gencerts clean

all: grom

grom: doc web cli
	cd cmd/grom && go build -o grom

cli:
	cd cmd/grom && go build -o grom

doc:
	cd api && swag init -d v1

web:
	cd ui/grom && flutter build web
	rm -rf internal/web/dist
	cp -r ui/grom/build/web internal/web/dist

android: android-apk

android-apk:
	cd ui/grom && flutter build apk --release

android-aab:
	cd ui/grom && flutter build appbundle --release

android-debug:
	cd ui/grom && flutter build apk --debug

gencerts:
	cd cmd/grom && go run . gencerts -ip $(IP) -domain $(DOMAIN)

clean:
	rm -f cmd/grom/grom
	cd ui/grom && flutter clean

.PHONY: all travka doc web cli android android-apk android-aab android-debug clean

all: travka

travka: doc web cli
	cd cmd/travka && go build -o travka

cli:
	cd cmd/travka && go build -o travka

doc:
	cd api && swag init -d v1

web:
	cd ui/travka && flutter build web
	rm -rf internal/web/dist
	cp -r ui/travka/build/web internal/web/dist

android: android-apk

android-apk:
	cd ui/travka && flutter build apk --release

android-aab:
	cd ui/travka && flutter build appbundle --release

android-debug:
	cd ui/travka && flutter build apk --debug

clean:
	rm -f cmd/travka/travka
	cd ui/travka && flutter clean

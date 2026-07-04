.PHONY: all travka travka-doc travka-web clean

all: travka

travka: travka-doc travka-web
	cd cmd/travka && go build -o travka

travka-doc:
	cd api && swag init -d v1

travka-web:
	cd ui/travka && flutter build web
	rm -rf internal/web/dist
	cp -r ui/travka/build/web internal/web/dist

clean:
	rm -f cmd/travka/travka
	cd ui/travka && flutter clean

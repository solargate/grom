.PHONY: all travka travka-doc travka-web clean

all: travka

travka: travka-doc travka-web
	cd cmd/travka && go build -o travka

travka-doc:
	cd api && swag init -d v1

travka-web:
	cd ui/travka && flutter build web

clean:
	rm -f cmd/travka/travka
	cd ui/travka && flutter clean

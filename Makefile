VERSION=$(shell git describe --tags 2>/dev/null|| echo "$$(git rev-parse HEAD)-$$(git rev-parse --abbrev-ref HEAD)")


.PHONY: build front version release                                                                                                                                                                     

release:
	make front
	bash build.sh $(VERSION)

version:
	@echo $(VERSION)

front:
	cd frontend && npm install
	cd frontend && npm run build

build:
	make front
	go build -o pingovalka .

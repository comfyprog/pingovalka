front:
	cd frontend && npm install
	cd frontend && npm run build

build:
	make front
	go build -o pingovalka .

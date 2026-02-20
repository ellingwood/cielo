.PHONY: build run test clean

build:
	go build -o bin/cielo ./cmd/cielo

run: build
	./bin/cielo

test:
	go test ./... -v -count=1

clean:
	rm -rf bin/

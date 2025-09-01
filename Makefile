BINARY = justify
CMD    = ./cmd/justify

build:
	mkdir -p dist
	go build -o dist/$(BINARY) $(CMD)

run: build
	./dist/$(BINARY)

clean:
	rm -rf dist

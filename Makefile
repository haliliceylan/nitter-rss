.PHONY: all build clean
build:
	go build -o build/ .
test:
	go test -v ./...
install-usr-local:
	make build
	sudo cp build/nitter-rss /usr/local/bin

build-multi-arch:
	./go-build-multiarch.sh github.com/haliliceylan/nitter-rss
build:
	@go build -o bin/blockchainPoker

run: build
	@./bin/blockchainPoker

test:
	@go test -v ./...

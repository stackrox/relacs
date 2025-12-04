cli:
	@mkdir -p bin
	@go build -o bin/relacs ./cmd/relacs

clean:
	@rm -f bin/relacs

unit-test:
	@go test ./... -v

lint:
	@golangci-lint run

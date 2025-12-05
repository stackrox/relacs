cli:
	@mkdir -p bin
	@go mod tidy
	@for GOOS in darwin linux; do \
		for GOARCH in amd64 arm64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -o bin/relacs-$$GOOS-$$GOARCH ./cmd/relacs; \
		done; \
	done

clean:
	@rm -f bin/relacs-*

unit-test:
	@go test ./... -v

lint:
	@golangci-lint run

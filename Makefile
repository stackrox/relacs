VERSION = $(shell git describe --tags)

LDFLAGS = -X 'github.com/stackrox/relacs/cmd/relacs/version.version=$(VERSION)'
cli:
	@mkdir -p bin
	@go mod tidy
	@for GOOS in darwin linux; do \
		for GOARCH in amd64 arm64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o bin/relacs-$$GOOS-$$GOARCH ./cmd/relacs; \
		done; \
	done

clean:
	@rm -f bin/relacs-*

unit-test:
	@go test ./cmd/... ./pkg/... -v

e2e-test:
	@go test ./tests/... -v

lint:
	@golangci-lint run

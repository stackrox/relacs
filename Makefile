VERSION = $(shell git describe --tags | sed -E 's/-[[:digit:]]+-g/+g/')

LDFLAGS = -X 'github.com/stackrox/relacs/cmd/relacs/version.version=$(VERSION)'
cli: clean
	@mkdir -p bin
	@go mod tidy
	@for GOOS in darwin linux; do \
		for GOARCH in amd64 arm64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o bin/relacs-$$GOOS-$$GOARCH ./cmd/relacs; \
		done; \
	done
	@sha256sum bin/relacs-* > bin/sha256sums.txt

clean:
	@rm -f bin/*

unit-test:
	@go test ./cmd/... ./pkg/... -v

e2e-test:
	@go test ./tests/... -v

lint:
	@golangci-lint run

version:
	@echo $(VERSION)

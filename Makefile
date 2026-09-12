.PHONY: build test test-integration vet lint fmt fmt-check tools tidy check

build:
	go build -o bin/qrocodile ./cmd/qrocodile

test:
	go test -count=1 ./...

test-integration:
	QR_API_INTEGRATION_REQUIRED=true go test -count=1 ./...

vet:
	go vet ./...

lint: tools
	./tools/bin/golangci-lint run ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

# tools is a phony alias for the real, file-based tools/bin/.stamp target below, so `make lint`
# only pays the `go install` cost when tools/go.mod or tools/go.sum actually changed.
tools: tools/bin/.stamp

# golangci-lint lives in tools/go.mod, a separate module — see CONTRIBUTING.md — so its own
# Go-version requirement never raises the floor this module's consumers need, and its huge
# transitive dependency tree never pollutes this module's go.mod/go.sum.
tools/bin/.stamp: tools/go.mod tools/go.sum
	cd tools && GOBIN="$(CURDIR)/tools/bin" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint
	@touch tools/bin/.stamp

tidy:
	go mod tidy
	cd tools && go mod tidy

check: fmt-check vet lint build test

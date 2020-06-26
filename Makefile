VERSION=2.0.0

GOLANGCI_VERSION=1.27.0
GOBINDATA_VERSION=3
GOX_VERSION=1.0.1

GO111MODULE=on
export GO111MODULE

.PHONY: all
all: deps assets build test lint

.PHONY: build
build: deps assets
	go build .
	cd cmd/mhsendmail && go build .

.PHONY: test
test: deps assets
	go test ./...

.PHONY: release
release: deps assets test lint
	gox -ldflags "-X main.version=${VERSION}" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

.PHONY: lint
lint: deps
	golangci-lint run

.PHONY: deps
deps:
	go mod download
	go get github.com/go-bindata/go-bindata/...@v${GOBINDATA_VERSION}
	go get github.com/golangci/golangci-lint@v${GOLANGCI_VERSION}
	go get github.com/mitchellh/gox@v${GOX_VERSION}

.PHONY: assets
assets: deps
	rm -f generated/assets/assets.go
	go-bindata -o generated/assets/assets.go -pkg assets assets/...

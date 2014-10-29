DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps bindata fmt
	go install

bindata:
	go-bindata assets/...

test: test-deps
	go list ./... | xargs -n1 go test

release: release-deps
	gox

fmt:
	go fmt ./...

deps:
	go get github.com/ian-kent/gotcha/...
	go get github.com/ian-kent/go-log/log
	go get github.com/ian-kent/envconf
	go get github.com/jteeuwen/go-bindata/...
	go get labix.org/v2/mgo

test-deps:
	go get github.com/stretchr/testify

release-deps:
	go get github.com/mitchellh/gox

.PNONY: all test deps bindata release fmt test-deps release-deps

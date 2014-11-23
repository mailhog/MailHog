DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps bindata fmt
	go install ./MailHog

bindata:
	go-bindata -o MailHog-UI/assets/assets.go -pkg assets MailHog-UI/assets/...

release: release-deps
	gox -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}"

fmt:
	go fmt ./...

deps:
	go get github.com/ian-kent/gotcha/...
	go get github.com/ian-kent/go-log/log
	go get github.com/ian-kent/envconf
	go get github.com/jteeuwen/go-bindata/...
	go get labix.org/v2/mgo

test-deps:
	go get github.com/smartystreets/goconvey

release-deps:
	go get github.com/mitchellh/gox

.PNONY: all bindata release fmt test-deps release-deps

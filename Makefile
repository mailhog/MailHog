DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps bindata fmt combined

combined:
	go install ./MailHog

server:
	go install ./MailHog-Server

ui:
	go install ./MailHog-UI

bindata:
	go-bindata -o MailHog-UI/assets/assets.go -pkg assets -prefix MailHog-UI/ MailHog-UI/assets/...

release: release-deps
	gox -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" ./MailHog

fmt:
	go fmt ./...

deps:
	go get github.com/ian-kent/gotcha/gotcha
	go get github.com/ian-kent/go-log/log
	go get github.com/ian-kent/envconf
	go get github.com/ian-kent/goose
	go get github.com/jteeuwen/go-bindata/...
	go get labix.org/v2/mgo
	# added to fix travis issues
	go get code.google.com/p/go-uuid/uuid
	go get code.google.com/p/go.crypto/bcrypt

test-deps:
	go get github.com/smartystreets/goconvey

release-deps:
	go get github.com/mitchellh/gox

.PNONY: all combined server ui bindata release fmt test-deps release-deps

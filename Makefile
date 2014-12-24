DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps fmt combined

combined:
	go install ./MailHog

release: release-deps
	gox -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" ./MailHog

fmt:
	go fmt ./...

deps:
	go get github.com/mailhog/MailHog-Server
	go get github.com/mailhog/MailHog-UI
	go get github.com/mailhog/http
	go get github.com/ian-kent/gotcha/gotcha
	go get github.com/ian-kent/go-log/log
	go get github.com/ian-kent/envconf
	go get github.com/ian-kent/goose
	go get github.com/ian-kent/linkio
	go get github.com/jteeuwen/go-bindata/...
	go get labix.org/v2/mgo
	# added to fix travis issues
	go get code.google.com/p/go-uuid/uuid
	go get code.google.com/p/go.crypto/bcrypt

test-deps:
	go get github.com/smartystreets/goconvey

release-deps:
	go get github.com/mitchellh/gox

.PNONY: all combined release fmt deps test-deps release-deps

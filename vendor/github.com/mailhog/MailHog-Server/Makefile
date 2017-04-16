DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: release-deps fmt combined

combined:
	go install .

release:
	gox -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

fmt:
	go fmt ./...

release-deps:
	go get github.com/mitchellh/gox

.PNONY: all combined release fmt release-deps

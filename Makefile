VERSION=1.0.0

all: fmt combined

combined:
	go install .

release: tag release-deps 
	gox -ldflags "-X main.version=${VERSION}" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}" .

fmt:
	go fmt ./...

release-deps:
	go get github.com/mitchellh/gox

pull:
	git pull
	cd ../data; git pull
	cd ../http; git pull
	cd ../MailHog-Server; git pull
	cd ../MailHog-UI; git pull
	cd ../smtp; git pull
	cd ../storage; git pull

tag:
	git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../data; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../http; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../MailHog-Server; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../MailHog-UI; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../smtp; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}
	cd ../storage; git tag -a -m 'v${VERSION}' v${VERSION} && git push origin v${VERSION}

.PHONY: all combined release fmt release-deps pull tag

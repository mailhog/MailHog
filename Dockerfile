FROM golang:1.7-alpine

RUN apk add --no-cache git \
	&& rm -rf /var/cache/apk/*

RUN go get github.com/mailhog/MailHog

EXPOSE 1025 8025

ENTRYPOINT ["/go/bin/MailHog"]

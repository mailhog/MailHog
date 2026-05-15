#
# MailHog Dockerfile
#

FROM golang:1.18-alpine AS builder

ENV GOPATH=/go
ENV GO111MODULE=off

# Place source at the expected import path
COPY . /go/src/github.com/mailhog/MailHog/
WORKDIR /go/src/github.com/mailhog/MailHog

# Build from local source using vendor directory:
RUN go build -o /usr/local/bin/mail-trapper .

FROM alpine:3
# Add mailhog user/group with uid/gid 1000.
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

COPY --from=builder /usr/local/bin/mail-trapper /usr/local/bin/mail-trapper

USER mailhog

WORKDIR /home/mailhog

CMD ["/usr/local/bin/mail-trapper"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025


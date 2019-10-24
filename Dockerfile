#
# MailHog Dockerfile
#
#  Build  with: docker build --tag mymongo .
#  Run with: docker run -d -p 8025:8025 -p 1025:1025 mymongo:latest

FROM golang:alpine AS builder

# Build MailHog
RUN apk --no-cache add --virtual build-dependencies git
RUN mkdir -p /root/gocode
RUN GOPATH=/root/gocode go get github.com/mailhog/MailHog

FROM alpine:latest

COPY --from=builder /root/gocode/bin/MailHog /usr/local/bin/

# Add mailhog user/group with uid/gid 1000.
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

USER mailhog

WORKDIR /home/mailhog

ENTRYPOINT ["/usr/local/bin/MailHog"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025

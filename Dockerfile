#
# MailHog Dockerfile
#

FROM golang:1.18-alpine AS builder

# Build MailHog from this checkout:
RUN apk --no-cache add --virtual build-dependencies \
    git

ENV GOPATH=/root/gocode
ENV GO111MODULE=off

WORKDIR /root/gocode/src/github.com/mailhog/MailHog
COPY . .

RUN go install .

FROM alpine:3
# Add mailhog user/group with uid/gid 1000.
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

COPY --from=builder /root/gocode/bin/MailHog /usr/local/bin/

USER mailhog

WORKDIR /home/mailhog

ENTRYPOINT ["MailHog"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025

#
# MailHog Dockerfile
#

FROM alpine:3.4

ENV HTTP_BASIC_AUTH_USER="" \
    HTTP_BASIC_AUTH_PASSWORD=""

# Install ca-certificates, required for the "release message" feature:
RUN apk --no-cache add \
    ca-certificates

# Install MailHog:
RUN apk --no-cache add --virtual build-dependencies \
    go \
    git \
  && mkdir -p /root/gocode \
  && export GOPATH=/root/gocode \
  && go get github.com/mailhog/MailHog \
  && mv /root/gocode/bin/MailHog /usr/local/bin \
  && rm -rf /root/gocode \
  && apk del --purge build-dependencies

# Add mailhog user/group with uid/gid 1000.
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

USER mailhog

WORKDIR /home/mailhog

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod a+x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025

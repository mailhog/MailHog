#
# MailHog Dockerfile
#

FROM golang:alpine

# Install MailHog:
RUN apk --no-cache add --virtual build-dependencies \
    git libcap \
  && mkdir -p /root/gocode \
  && export GOPATH=/root/gocode \
  && go get github.com/mailhog/MailHog \
  && mv /root/gocode/bin/MailHog /usr/local/bin \
  && rm -rf /root/gocode \
  && setcap  CAP_NET_BIND_SERVICE=+eip /usr/local/bin/MailHog \
  && apk del --purge build-dependencies libcap 

# Add mailhog user/group with uid/gid 1000.
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

USER mailhog

WORKDIR /home/mailhog

ENTRYPOINT ["MailHog"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025

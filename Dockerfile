#
# MailHog Dockerfile
#

FROM debian:buster-slim AS builder
SHELL ["/bin/bash", "-ec"]

# Install build-time dependancies
RUN apt-get -yq update
RUN apt-get -yq install --no-install-{recommends,suggests} curl ca-certificates git make build-essential

# Install go
ARG GO_VERSION=1.14.4
RUN tar -xzC / < <(curl -f -s https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz)

# Create build-user user
ARG BUILD_USERNAME=build-user
ARG BUILD_DIRECTORY=/home/build-user
RUN useradd -md ${BUILD_DIRECTORY} -u 1000 ${BUILD_USERNAME}

COPY . ${BUILD_DIRECTORY}/MailHog
RUN find ${BUILD_DIRECTORY}/MailHog -exec ls -lshd '{}' +
RUN chown -R ${BUILD_USERNAME}:${BUILD_USERNAME} ${BUILD_DIRECTORY}/MailHog

# Build MailHog
USER ${BUILD_USERNAME}
WORKDIR ${BUILD_DIRECTORY}
ENV GOPATH="${BUILD_DIRECTORY}/go"
ENV PATH="$PATH:/go/bin:${GOPATH}/bin"
RUN mkdir -p go/{src,bin} bin
RUN make -C MailHog deps
RUN make -C MailHog
RUN mv MailHog/MailHog MailHog/cmd/mhsendmail/mhsendmail ${BUILD_DIRECTORY}/bin

FROM debian:buster-slim

# Create mailhog user as non-login system user with user-group
ARG USERNAME=mailhog
RUN useradd --shell /bin/false -Urb / -u 99 ${USERNAME}

# Copy mailhog binary
COPY --from=builder /home/build-user/bin/* /bin/

# Expose the SMTP and HTTP ports:
EXPOSE 2525 8025

WORKDIR /
USER ${USERNAME}
CMD ["MailHog"]

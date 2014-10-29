FROM ubuntu:14.04

EXPOSE 1025
EXPOSE 8025

RUN apt-get update -qq
RUN apt-get install -qqy ca-certificates

ADD Go-MailHog /tmp/

CMD ["./tmp/Go-MailHog"]

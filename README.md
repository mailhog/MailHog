Go-MailHog
=========

Inspired by [MailCatcher](http://mailcatcher.me/), easier to install.

Go-MailHog is a rewritten version of [MailHog](https://github.com/ian-kent/MailHog), which was born out of [M3MTA](https://github.com/ian-kent/M3MTA).

Go was chosen for portability - MailHog runs without installation on multiple platforms.

### Requirements

* None!

### Getting started

* Download the latest release of Go-MailHog for your platform
* Start MailHog

By default, the SMTP server will start on port 1025, and the HTTP
server will start on port 8025.

### Features

* ESMTP server implementing RFC5321
* Support for SMTP AUTH (RFC4954) and PIPELINING (RFC2920)
* Web interface to view messages (plain text, HTML or source)
* Release messages to real SMTP servers
* HTTP API to list, retrieve and delete messages
  * See (APIv1 documentation)[APIv1.md] for more information
* Multipart MIME support
* Download individual MIME parts
* In-memory message storage
* MongoDB storage for message persistence
* Lightweight and portable
* No installation required

![Screenshot of MailHog web interface](/images/MailHog.png "MailHog web interface")

### Configuration

You can configure Go-MailHog using command line options:

| Parameter     | Default         | Description
| ------------- | --------------- | -----------
| -hostname     | mailhog.example | Hostname to use for EHLO/HELO and message IDs
| -httpbindaddr | 0.0.0.0:8025    | Interface and port for HTTP server to bind to
| -mongocoll    | messages        | MongoDB collection name for message storage
| -mongodb      | mailhog         | MongoDB database name for message storage
| -mongouri     | 127.0.0.1:27017 | MongoDB host and port
| -smtpbindaddr | 0.0.0.0:1025    | Interface and port for SMTP server to bind to
| -storage      | memory          | Set message storage: memory / mongodb

### Contributing

Clone this repository to ```$GOPATH/src/github.com/ian-kent/MailHog``` and type ```go install```.

You'll need go-bindata to embed web assets:
```
go get https://github.com/jteeuwen/go-bindata/...
go-bindata assets/...
```

Run tests using ```go test```. You'll need a copy of MailHog running for tests to pass.
(Tests currently fail using in-memory storage, use MongoDB!)

If you make any changes, run ```go fmt``` before submitting a pull request.

### Licence

Copyright ©‎ 2014, Ian Kent (http://www.iankent.eu).

Released under MIT license, see [LICENSE](license) for details.

Go-MailHog
=========

Inspired by [MailCatcher](http://mailcatcher.me/), easier to install.

Go-MailHog is a rewritten version of [MailHog](https://github.com/ian-kent/MailHog), which was born out of [M3MTA](https://github.com/ian-kent/M3MTA).

Go was chosen for portability - MailHog runs without installation on multiple platforms.

### Requirements

* None!
* Well, you need MongoDB installed somewhere

### Getting started

* Download the latest release of Go-MailHog for your platform
* Start MailHog

By default, the SMTP server will start on port 1025, and the HTTP
server will start on port 8025.

### Features

* ESMTP server implementing RFC5321
* Web interface to view messages
* API interface to list, retrieve and delete messages
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

### Contributing

Clone this repository to ```$GOPATH/src/github.com/ian-kent/MailHog``` and type ```go install```.

Run tests using ```go test```. You'll need a copy of MailHog running for tests to pass.

If you make any changes, run ```go fmt``` before submitting a pull request.

### Licence

Copyright ©‎ 2014, Ian Kent (http://www.iankent.eu).

Released under MIT license, see [LICENSE](license) for details.

MailHog [ ![Download](https://img.shields.io/github/release/mailhog/MailHog.svg) ](https://github.com/mailhog/MailHog/releases/tag/v1.0.0) [![GoDoc](https://godoc.org/github.com/mailhog/MailHog?status.svg)](https://godoc.org/github.com/mailhog/MailHog) [![Build Status](https://travis-ci.org/mailhog/MailHog.svg?branch=master)](https://travis-ci.org/mailhog/MailHog)
=========

Inspired by [MailCatcher](http://mailcatcher.me/), easier to install.

* Download and run MailHog
* Configure your outgoing SMTP server
* View your outgoing email in a web UI
* Release it to a real mail server

Built with Go - MailHog runs without installation on multiple platforms.

### Overview

MailHog is an email testing tool for developers:

* Configure your application to use MailHog for SMTP delivery
* View messages in the web UI, or retrieve them with the JSON API
* Optionally release messages to real SMTP servers for delivery

### Installation

#### Manual installation
[Download the latest release for your platform](/docs/RELEASES.md). Then
[read the deployment guide](/docs/DEPLOY.md) for deployment options.

#### MacOS
```bash
brew update && brew install mailhog
```

Then, start MailHog by running `mailhog` in the command line.

#### Debian / Ubuntu
```bash
sudo apt-get -y install golang-go
go get github.com/mailhog/MailHog
```

Then, start MailHog by running `/path/to/MailHog` in the command line.

E.g. the path to Go's bin files on Ubuntu is `~/go/bin/`, so to start the MailHog run:

```bash
~/go/bin/MailHog
```

### Configuration

Check out how to [configure MailHog](/docs/CONFIG.md), or use the default settings:
  * the SMTP server starts on port 1025
  * the HTTP server starts on port 8025
  * in-memory message storage

### Features

See [MailHog libraries](docs/LIBRARIES.md) for a list of MailHog client libraries.

* ESMTP server implementing RFC5321
* Support for SMTP AUTH (RFC4954) and PIPELINING (RFC2920)
* Web interface to view messages (plain text, HTML or source)
  * Supports RFC2047 encoded headers
* Real-time updates using EventSource
* Release messages to real SMTP servers
* Chaos Monkey for failure testing
  * See [Introduction to Jim](/docs/JIM.md) for more information
* HTTP API to list, retrieve and delete messages
  * See [APIv1](/docs/APIv1.md) and [APIv2](/docs/APIv2.md) documentation for more information
* [HTTP basic authentication](docs/Auth.md) for MailHog UI and API
* Multipart MIME support
* Download individual MIME parts
* In-memory message storage
* MongoDB and file based storage for message persistence
* Lightweight and portable
* No installation required

#### sendmail

[mhsendmail](https://github.com/mailhog/mhsendmail) is a sendmail replacement for MailHog.

It redirects mail to MailHog using SMTP.

You can also use `MailHog sendmail ...` instead of the separate mhsendmail binary.

Alternatively, you can use your native `sendmail` command by providing `-S`, for example:

```bash
/usr/sbin/sendmail -S mail:1025
```

For example, in PHP you could add either of these lines to `php.ini`:

```
sendmail_path = /usr/local/bin/mhsendmail
sendmail_path = /usr/sbin/sendmail -S mail:1025
```

#### Web UI

![Screenshot of MailHog web interface](/docs/MailHog.png "MailHog web interface")

### Contributing

MailHog is a rewritten version of [MailHog](https://github.com/ian-kent/MailHog), which was born out of [M3MTA](https://github.com/ian-kent/M3MTA).

Clone this repository to ```$GOPATH/src/github.com/mailhog/MailHog``` and type ```make deps```.

See the [Building MailHog](/docs/BUILD.md) guide.

Requires Go 1.4+ to build.

Run tests using ```make test``` or ```goconvey```.

If you make any changes, run ```go fmt ./...``` before submitting a pull request.

### Licence

Copyright ©‎ 2014 - 2017, Ian Kent (http://iankent.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

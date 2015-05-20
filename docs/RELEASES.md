MailHog Releases
================

### [v0.1.7](https://github.com/mailhog/MailHog/releases/v0.1.7)
- Add [mhsendmail](https://github.com/mailhog/mhsendmail) sendmail replacement
- Fix #42 - panic when setting UI bind address
- Fix #46 - utf8 error in e-mail subject
- Fix #41 and #50 - underscores replaced with spaces in UI
- Fix mailhog/MailHog-UI#6 - user defined pagination
- Merge #43 and #44 - fix documentation, thanks @eirc
- Merge #48 - fix documentation, thanks @zhubert
- Merge mailhog/MailHog-Server#1 - avoid duplicate headers, thanks @wienczny

### [v0.1.6](https://github.com/mailhog/MailHog/releases/v0.1.6)
- Fix #24 - base64 attachments/mime part downloads
- Fix #28 - embed js/css/font assets for offline use
- Fix #29 - overview of MailHog for readme
- Fix #34 - message list scrolling
- Fix #35 - message list sorting
- Fix #36 - document outgoing SMTP server configuration and APIv2
- Merge mailhog/MailHog-UI#4 - support base64 content transfer encoding, thanks @stekershaw
- Merge mailhog/Mailhog-UI#5 - single part encoded text/plain, thanks @naoina

### [v0.1.5](https://github.com/mailhog/MailHog/releases/v0.1.5)
- Fix mailhog/MailHog-UI#3 - squashed subject line

### [v0.1.4](https://github.com/mailhog/MailHog/releases/v0.1.4)
- Merge mailhog/data#2 - MIME boundary fixes, thanks @nvcnvn
- Merge mailhog/MailHog-UI#2 - UI overhaul, thanks @omegahm
- Fix #31 - updated this file :smile:

### [v0.1.3](https://github.com/mailhog/MailHog/releases/v0.1.3)
- Fix #22 - render non-multipart messages with HTML content type
- Fix #25 - make web UI resource paths relative

### [v0.1.2](https://github.com/mailhog/MailHog/releases/v0.1.2)
- Hopefully fix #22 - broken rendering of HTML email
- Partially implement #15 - authentication for SMTP release
  - Load outgoing SMTP servers from file
  - Save outgoing SMTP server when releasing message in UI
  - Select outgoing SMTP server when release message in UI
- Make Jim (Chaos Monkey) available via APIv2
- Add Jim overview and on/off switch to web UI

### [v0.1.1](https://github.com/mailhog/MailHog/releases/v0.1.1)
- Fix #23 - switch to iframe to fix CSS bug
- Update to latest AngularJS
- Update Dockerfile - thanks @humboldtux
- Fix SMTP AUTH bug (missing from EHLO)
- Fix SMTP new line parsing

### [v0.1.0](https://github.com/mailhog/MailHog/releases/v0.1.0)

- Switch to semantic versioning
- Rewrite web user interface
- Deprecate APIv1
- Rewrite messages endpoint for APIv2
- Add search to APIv2

### [v0.09](https://github.com/mailhog/MailHog/releases/0.08)

- Fix #8 - add Chaos Monkey ([Jim](JIM.md)) to support failure testing

### [v0.08](https://github.com/mailhog/MailHog/releases/0.08)

- Extract SMTP protocol into isolated library
- Better protocol tests
- Add hooks for manipulating protocol behaviour
- Merge #14 - fix display of multipart messges, thanks @derwassi
- Merge #17 - fix API v1 docs, thanks @geerlingguy
- Fix #11 - add build documentation
- Fix #12 - support broken MAIL/RCPT syntax
- Fix #16 - add deployment documentation
- Fix #18 - better server-sent event support using [goose](https://github.com/ian-kent/goose)

### [v0.07](https://github.com/mailhog/MailHog/releases/tag/0.07)

- Fix #6 - Make SMTP verbs case-insensitive

### [v0.06](https://github.com/mailhog/MailHog/releases/tag/0.06)

- Fix #5 - Support leading tab in multiline headers

### [v0.05](https://github.com/mailhog/MailHog/releases/tag/0.05)

- Add #4 - UI support for RFC2047 encoded headers

### [v0.04](https://github.com/mailhog/MailHog/releases/tag/0.04)

* Configure from environment
* Include example Dockerfile
* Fix #1 - mismatched import path and repository name
* Fix #2 - possible panic with some MIME content
* Fix #3 - incorrect handling of RSET


### [v0.03](https://github.com/mailhog/MailHog/releases/tag/0.03)

* Download message in .eml format
* Cleaned up v1 API
* Web UI and API improvements
  * Fixed UI rendering bugs
  * Message search and matched/total message count
  * Message list resizing and scrolling  
  * EventSource support for message streaming
  * Better error handling and reporting
  * View/download individual MIME parts
  * Release messages to real SMTP servers
* Switch to [go-bindata](https://github.com/jteeuwen/go-bindata) for asset embedding

### [v0.02](https://github.com/mailhog/MailHog/releases/tag/0.02)

* Better support for ESMTP (RFC5321)
* Support for SMTP AUTH (RFC4954) and PIPELINING (RFC2920)
* Improved AJAX web interface to view messages (plain text, HTML or source)
* Improved HTTP API to list, retrieve and delete messages
* Multipart MIME support
* In-memory message storage
* MongoDB storage for message persistence

### [v0.01](https://github.com/mailhog/MailHog/releases/tag/0.01)

* Basic support for SMTP and HTTP servers
* Accepts SMTP messages
* Stores parsed messages in MongoDB
* Makes messages available via API
* has Bootstrap/AngularJS UI for viewing/deleting messages

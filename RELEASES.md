Go-MailHog Releases
===================

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

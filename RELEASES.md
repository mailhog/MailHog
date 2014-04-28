Go-MailHog Releases
===================

### [v0.03](https://github.com/ian-kent/Go-MailHog/releases/tag/0.03)

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

### [v0.02](https://github.com/ian-kent/Go-MailHog/releases/tag/0.02)

* Better support for ESMTP (RFC5321)
* Support for SMTP AUTH (RFC4954) and PIPELINING (RFC2920)
* Improved AJAX web interface to view messages (plain text, HTML or source)
* Improved HTTP API to list, retrieve and delete messages
* Multipart MIME support
* In-memory message storage
* MongoDB storage for message persistence

### [v0.01](https://github.com/ian-kent/Go-MailHog/releases/tag/0.01)

* Basic support for SMTP and HTTP servers
* Accepts SMTP messages
* Stores parsed messages in MongoDB
* Makes messages available via API
* has Bootstrap/AngularJS UI for viewing/deleting messages

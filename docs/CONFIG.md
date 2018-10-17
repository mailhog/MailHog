Configuring MailHog
===================

You can configure MailHog using command line options or environment variables:

| Environment         | Command line    | Default         | Description
| ------------------- | --------------- | --------------- | -----------
| MH_CORS_ORIGIN      | -cors-origin    |                 | If set, a Access-Control-Allow-Origin header is returned for API endpoints
| MH_HOSTNAME         | -hostname       | mailhog.example | Hostname to use for EHLO/HELO and message IDs
| MH_API_BIND_ADDR    | -api-bind-addr  | 0.0.0.0:8025    | Interface and port for HTTP API server to bind to
| MH_UI_BIND_ADDR     | -ui-bind-addr   | 0.0.0.0:8025    | Interface and port for HTTP UI server to bind to
| MH_MAILDIR_PATH     | -maildir-path   |                 | Maildir path (for maildir storage backend)
| MH_MONGO_COLLECTION | -mongo-coll     | messages        | MongoDB collection name for message storage
| MH_MONGO_DB         | -mongo-db       | mailhog         | MongoDB database name for message storage
| MH_MONGO_URI        | -mongo-uri      | 127.0.0.1:27017 | MongoDB host and port
| MH_SMTP_BIND_ADDR   | -smtp-bind-addr | 0.0.0.0:1025    | Interface and port for SMTP server to bind to
| MH_STORAGE          | -storage        | memory          | Set message storage: memory / mongodb / maildir
| MH_OUTGOING_SMTP    | -outgoing-smtp  |                 | JSON file defining outgoing SMTP servers
| MH_UI_WEB_PATH      | -ui-web-path    |                 | WebPath under which the ui is served (without leading or trailing slahes), e.g. 'mailhog'
| MH_AUTH_FILE        | -auth-file      |                 | A username:bcryptpw mapping file

#### Note on HTTP bind addresses

If `api-bind-addr` and `ui-bind-addr` are identical, a single listener will
be used allowing both to co-exist on one port.

The values must match in a string comparison. Resolving to the same host and
port combination isn't enough.

### Outgoing SMTP configuration

Outgoing SMTP servers can be set in web UI when releasing a message, and can
be temporarily persisted for later use in the same session.

To make outgoing SMTP servers permanently available, create a JSON file with
the following structure, and set `MH_OUTGOING_SMTP` or `-outgoing-smtp`.

```json
{
    "server name": {
        "name": "server name",
        "host": "...",
        "port": "587",
        "email": "...",
        "username": "...",
        "password": "...",
        "mechanism": "PLAIN"
    }
}
```

Only `name`, `host` and `port` are required.

`mechanism` can be `PLAIN` or `CRAM-MD5`.

### Firewalls and proxies

If you have MailHog behind a firewall, you'll need ports `8025` and `8080` by default.

You can override this using `-api-bind-addr`, `-ui-bind-addr` and `-smtp-bind-addr` configuration options.

If you're using MailHog behind a reverse proxy, e.g. nginx, make sure WebSocket connections
are also supported and configured - see [this issue](https://github.com/mailhog/MailHog/issues/117) for information.

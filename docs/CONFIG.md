Configuring MailHog
===================

You can configure MailHog using command line options or environment variables:

| Environment         | Command line    | Default         | Description
| ------------------- | --------------- | --------------- | -----------
| MH_CORS_ORIGIN      | -cors-origin    |                 | If set, a Access-Control-Allow-Origin header is returned for API endpoints
| MH_HOSTNAME         | -hostname       | mailhog.example | Hostname to use for EHLO/HELO and message IDs
| MH_API_BIND_ADDR    | -api-bind-addr  | 0.0.0.0:8025    | Interface and port for HTTP UI server to bind to
| MH_UI_BIND_ADDR     | -ui-bind-addr   | 0.0.0.0:8025    | Interface and port for HTTP API server to bind to
| MH_MONGO_COLLECTION | -mongo-coll     | messages        | MongoDB collection name for message storage
| MH_MONGO_DB         | -mongo-db       | mailhog         | MongoDB database name for message storage
| MH_MONGO_URI        | -mongo-uri      | 127.0.0.1:27017 | MongoDB host and port
| MH_SMTP_BIND_ADDR   | -smtp-bind-addr | 0.0.0.0:1025    | Interface and port for SMTP server to bind to
| MH_STORAGE          | -storage        | memory          | Set message storage: memory / mongodb
| MH_OUTGOING_SMTP    | -outgoing-smtp  |                 | JSON file defining outgoing SMTP servers

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

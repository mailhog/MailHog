Configuring MailHog
===================

You can configure MailHog using command line options or environment variables:

| Environment         | Command line  | Default         | Description
| ------------------- | ------------- | --------------- | -----------
| MH_CORS_ORIGIN      | -cors-origin  |                 | If set, a Access-Control-Allow-Origin header is returned for API endpoints
| MH_HOSTNAME         | -hostname     | mailhog.example | Hostname to use for EHLO/HELO and message IDs
| MH_HTTP_BIND_ADDR   | -httpbindaddr | 0.0.0.0:8025    | Interface and port for HTTP server to bind to
| MH_MONGO_COLLECTION | -mongocoll    | messages        | MongoDB collection name for message storage
| MH_MONGO_DB         | -mongodb      | mailhog         | MongoDB database name for message storage
| MH_MONGO_URI        | -mongouri     | 127.0.0.1:27017 | MongoDB host and port
| MH_SMTP_BIND_ADDR   | -smtpbindaddr | 0.0.0.0:1025    | Interface and port for SMTP server to bind to
| MH_STORAGE          | -storage      | memory          | Set message storage: memory / mongodb

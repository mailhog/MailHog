Authentication
==============

HTTP basic authentication is supported using a password file.

See [example-auth](example-auth) for an example (the password is `test`).

Authentication applies to all HTTP requests, including static content
and API endpoints.

### Password file format

The password file format is:

* One user per line
* `username:password`
* Password is bcrypted

By default, a bcrypt difficulty of 4 is used to reduce page load times.

### Generating a bcrypted password

You can use a MailHog shortcut to generate a bcrypted password:

    MailHog bcrypt <password>

### Enabling HTTP authentication

To enable authentication, pass an `-auth-file` flag to MailHog:

    MailHog -auth-file=docs/example-auth

This also works if you're running MailHog-UI and MailHog-Server separately:

    MailHog-Server -auth-file=docs/example-auth
    MailHog-UI -auth-file=docs/example-auth

## Future compatibility

Authentication has been a bit of an experiment.

The exact implementation may change over time, e.g. using sessions in the UI
and tokens for the API to avoid frequently bcrypting passwords.

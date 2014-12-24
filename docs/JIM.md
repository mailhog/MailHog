Introduction to Jim
===================

Jim is the MailHog Chaos Monkey, inspired by Netflix.

You can invite Jim to the party using the `invite-jim` flag:

    MailHog -invite-jim

With Jim around, things aren't going to work how you expect.

### What can Jim do?

* Reject connections
* Rate limit connections
* Reject authentication
* Reject senders
* Reject recipients

It does this randomly, but within defined parameters.

You can control these using the following command line flags:

| Flag                  | Default | Description
| --------------------- | ------- | ----
| -invite-jim           | false   | Set to true to invite Jim
| -jim-disconnect       | 0.005   | Chance of randomly disconnecting a session
| -jim-accept           | 0.99    | Chance of accepting an incoming connection
| -jim-linkspeed-affect | 0.1     | Chance of applying a rate limit
| -jim-linkspeed-min    | 1024    | Minimum link speed (in bytes per second)
| -jim-linkspeed-max    | 10240   | Maximum link speed (in bytes per second)
| -jim-reject-sender    | 0.05    | Chance of rejecting a MAIL FROM command
| -jim-reject-recipient | 0.05    | Chance of rejecting a RCPT TO command
| -jim-reject-auth      | 0.05    | Chance of rejecting an AUTH command

If you enable Jim, you enable all parts. To disable individual parts, set the chance
of it happening to 0, e.g. to disable connection rate limiting:

    MailHog -invite-jim -jim-linkspeed-affect=0

### Examples

Always rate limit to 1 byte per second:

    MailHog -invite-jim -jim-linkspeed-affect=1 -jim-linkspeed-max=1 -jim-linkspeed-min=1

Disconnect clients after approximately 5 commands:

    MailHog -invite-jim -jim-disconnect=0.2

Simulate a mobile connection (at 10-100kbps) for 10% of clients:

    MailHog -invite-jim -jim-linkspeed-affect=0.1 -jim-linkspeed-min=1250 -jim-linkspeed-max=12500

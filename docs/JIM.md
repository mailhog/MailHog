Introduction to Jim
===================

Jim is the MailHog Chaos Monkey, inspired by Netflix.

You can invite Jim to the party using the `invite-jim` flag:

    MailHog -invite-jim

Or by making a request to [Jim's API](./APIv2.md):

    curl -X POST http://example.com:8025/api/v2/jim
  
With Jim around, things aren't going to work how you expect.

### What can Jim do?

* Reject connections
* Rate limit connections
* Reject authentication
* Reject senders
* Reject recipients

It does this randomly, but within defined parameters.

## Control Jim from the command line

You can control Jims behavior using the following command line flags:

| Flag                  | Default | Description
| --------------------- | ------- | ----
| -invite-jim           | true    | Set to true to invite Jim
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

### Command line examples

Always rate limit to 1 byte per second:

    MailHog -invite-jim -jim-linkspeed-affect=1 -jim-linkspeed-max=1 -jim-linkspeed-min=1

Disconnect clients after approximately 5 commands:

    MailHog -invite-jim -jim-disconnect=0.2

Simulate a mobile connection (at 10-100kbps) for 10% of clients:

    MailHog -invite-jim -jim-linkspeed-affect=0.1 -jim-linkspeed-min=1250 -jim-linkspeed-max=12500

## Control Jim from the API

You can control Jim's behavior using [the API](./APIv2.md) at `/api/v2/jim`.

The API accepts a JSON payload with the following properties:

| Property              | Default  | Description
| --------------------- | -------- | ----
| DisconnectChance      | 0.005    | Chance of randomly disconnecting a session
| AcceptChance          | 0.99     | Chance of accepting an incoming connection
| LinkSpeedAffect       | 0.1      | Chance of applying a rate limit
| LinkSpeedMin          | 1024     | Minimum link speed (in bytes per second)
| LinkSpeedMax          | 10240    | Maximum link speed (in bytes per second)
| RejectSenderChance    | 0.05     | Chance of rejecting a MAIL FROM command
| RejectRecipientChance | 0.05     | Chance of rejecting a RCPT TO command
| RejectAuthChance      | 0.05     | Chance of rejecting an AUTH command

The default values are used only when Jim is enabled without providing a JSON payload. 
When a JSON payload is sent to the API, any properties that are not provided are set to 0.

When Jim sees this:

```json
{
    "DisconnectChance": 1.0
}
```

It assumes you meant this:

```json
{
    "DisconnectChance": 1.0,
    "AcceptChance": 0,
    "LinkSpeedAffect": 0,
    "LinkSpeedMin": 0,
    "LinkSpeedMax": 0,
    "RejectSenderChance": 0,
    "RejectRecipientChance": 0,
    "RejectAuthChance": 0
}
```

### API examples

Invite Jim with with a specific configuration:

    curl -X POST -d '{"DisconnectChance": 1.0}' http://example.com:8025/api/v2/jim

Update Jim's configuration on the fly:

    curl -X PUT -d '{"AcceptChance": 0.1}' http://example.com:8025/api/v2/jim

Tell Jim the party is over:

    curl -X DELETE http://example.com:8025/api/v2/jim

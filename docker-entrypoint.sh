#!/usr/bin/env sh

if [ -n "$HTTP_BASIC_AUTH_USER" ] && [ -n "$HTTP_BASIC_AUTH_PASSWORD" ]
then
    echo "Generating password file: /home/mailhog/htpasswd"
    echo "$HTTP_BASIC_AUTH_USER:$(MailHog bcrypt $HTTP_BASIC_AUTH_PASSWORD)" > /home/mailhog/htpasswd
    MailHog -auth-file=/home/mailhog/htpasswd
else
    MailHog
fi

exit 0

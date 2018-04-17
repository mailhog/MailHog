#!/bin/sh
# shellcheck shell=dash

#
# Sends email to the given SMTP server via Netcat.
#
# Usage:
# echo 'Text' | ./mail.sh [-h host] [-p port] [-f from] [-t to] [-s subject]
#
# Copyright 2016, Sebastian Tschan
# https://blueimp.net
#
# Licensed under the MIT license:
# https://opensource.org/licenses/MIT
#

set -e

# Default settings:
HOST=localhost
PORT=25
USER=${USER:-user}
# shellcheck disable=SC2169
HOSTNAME=${HOSTNAME:-localhost}
FROM="$USER <$USER@$HOSTNAME>"
TO='test <test@example.org>'
SUBJECT=Test

NEWLINE='
'

print_usage() {
  echo \
    "Usage: echo 'Text' | $0 [-h host] [-p port] [-f from] [-t to] [-s subject]"
}

# Prints the given error and optionally a usage message and exits:
error_exit() {
  echo "Error: $1" >&2
  if [ ! -z "$2" ]; then
    print_usage >&2
  fi
  exit 1
}

# Adds brackets around the last word in the given address, trims whitespace:
normalize_address() {
  local address
  address=$(echo "$1" | awk '{$1=$1};1')
  if [ "${address%>}" = "$address" ]; then
    echo "$address" | sed 's/[^ ]*$/<&>/'
  else
    echo "$address"
  fi
}

# Does a simple validity check on the email address format,
# without support for comments or for quoting in the local-part:
validate_email() {
  local local_part=${1%%@*>}
  local_part=$(echo "${local_part#<}" | sed 's/[][[:cntrl:][:space:]"(),:;\]//')
  local domain=${1##<*@}
  domain=$(echo "${domain%>}" | LC_CTYPE=UTF-8 sed 's/[^][[:alnum:].:-]//')
  if [ "<$local_part@$domain>" != "$1" ]; then
    error_exit "Invalid email address: $1"
  fi
}

is_printable_ascii() {
  (LC_CTYPE=C; case "$1" in *[![:print:]]*) return 1;; esac)
}

# Encodes the given string according to RFC 1522:
# https://tools.ietf.org/html/rfc1522
rfc1342_encode() {
  if is_printable_ascii "$1"; then
    printf %s "$1"
  else
    printf '=?utf-8?B?%s?=' "$(printf %s "$1" | base64)"
  fi
}

encode_address() {
  local email="<${1##*<}"
  if [ "$email" != "$1" ]; then
    local name="${1%<*}"
    # Remove any trailing space as we add it again in the next line:
    name="${name% }"
    echo "$(rfc1342_encode "$name") $email"
  else
    echo "$1"
  fi
}

parse_recipients() {
  local addresses
  local address
  local email
  local output
  local recipients
  addresses=$(echo "$TO" | tr ',' '\n')
  IFS="$NEWLINE"
  for address in $addresses; do
    address=$(normalize_address "$address")
    email="<${address##*<}"
    validate_email "$email"
    output="$output, $(encode_address "$address")"
    recipients="$recipients$NEWLINE$email"
  done
  unset IFS
  # Remove the first commma and space from the address list:
  TO="$(echo "$output" | cut -c 3-)"
  # Remove leading blank line from the recipients list and add header prefixes:
  RECIPIENTS_HEADERS="$(echo "$recipients" | sed '/./,$!d; s/^/RCPT TO:/')"
}

parse_sender() {
  local email
  FROM="$(normalize_address "$FROM")"
  email="<${FROM##*<}"
  validate_email "$email"
  FROM="$(encode_address "$FROM")"
  SENDER_HEADER="MAIL FROM:$email"
}

parse_text() {
  local line
  CONTENT_TRANSFER_ENCODING=7bit
  TEXT=
  while read -r line; do
    # Use base64 encoding if the text contains non-printable ASCII characters
    # or exceeds 998 characters (excluding the \r\n line endings):
    if ! is_printable_ascii "$line" || [ "${#line}" -gt 998 ]; then
      CONTENT_TRANSFER_ENCODING=base64
    fi
    TEXT="$TEXT$line$NEWLINE"
  done
  if [ "$CONTENT_TRANSFER_ENCODING" = base64 ]; then
    TEXT="$(printf %s "$TEXT" | base64)"
  else
    # Prepend each period at the start of a line with another period,
    # to follow RFC 5321 Section 4.5.2 Transparency guidelines:
    TEXT="$(printf %s "$TEXT" | sed 's/^\./.&/g')"
  fi
}

while getopts ':h:p:f:t:s:' OPT; do
  case "$OPT" in
    h)  HOST="$OPTARG";;
    p)  PORT="$OPTARG";;
    f)  FROM="$OPTARG";;
    t)  TO="$OPTARG";;
    s)  SUBJECT="$OPTARG";;
    :)  error_exit "Option -$OPTARG requires an argument." true;;
    \?) error_exit "Invalid option: -$OPTARG" true;;
  esac
done

parse_recipients
parse_sender
parse_text

SUBJECT="$(rfc1342_encode "$SUBJECT")"
DATE=$(date '+%a, %d %b %Y %H:%M:%S %z')

MAIL='HELO '"$HOSTNAME"'
'"$SENDER_HEADER"'
'"$RECIPIENTS_HEADERS"'
DATA
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: '"$CONTENT_TRANSFER_ENCODING"'
Date: '"$DATE"'
From: '"$FROM"'
To: '"$TO"'
Subject: '"$SUBJECT"'

'"$TEXT"'
.
QUIT'

echo "$MAIL" | awk '{printf "%s\r\n", $0}' | nc "$HOST" "$PORT"

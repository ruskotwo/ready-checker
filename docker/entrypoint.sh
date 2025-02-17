#!/bin/sh

echo "$@"

#[For Dev]
#/go/bin/dlv --listen=:40001 --continue --accept-multiclient --headless=true --api-version=2 exec /usr/local/bin/ready-checker "$@"

exec /usr/local/bin/ready-checker "$@"

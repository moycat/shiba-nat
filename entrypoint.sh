#!/bin/bash

case "$1" in
  gateway)
    exec /usr/bin/shiba-nat-gateway
    ;;
  client)
    exec /usr/bin/shiba-nat-client
    ;;
  *)
    exec "$@"
    ;;
esac

#!/bin/bash

ROLE="registry,server"

IFS=','

read -r -a roles <<<"$ROLE"

for role in "${roles[@]}"; do
  case "$role" in
  "registry")
    ximager registry /etc/ximager/registry.yaml --config /etc/ximager/ximager.yaml &
    ;;
  "server")
    ximager server --config /etc/ximager/ximager.yaml &
    ;;
  *)
    echo "Unknown role: $role"
    exit 1
    ;;
  esac
done

tail -f /dev/null

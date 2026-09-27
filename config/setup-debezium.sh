#!/bin/bash

for i in $(seq 1 90); do
  curl -sf http://localhost:8083/connectors >/dev/null 2>&1 && break
  echo "Waiting for connect..."; sleep 2
done

curl -sf http://localhost:8083/connectors >/dev/null 2>&1 \
  || { echo "Connect REST API never came up" >&2; exit 1; }
curl -sf -X PUT -H 'Content-Type: application/json' \
  --data-binary @"$(dirname "$0")/debezium/outbox-connector.json" \
  http://localhost:8083/connectors/outbox-connector/config

echo "Outbox connector registered"
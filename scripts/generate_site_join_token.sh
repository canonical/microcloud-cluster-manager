#!/usr/bin/env bash
set -e
# Execute the curl command with the provided values
curl --insecure -H 'Content-Type: application/json' -d '{
  "expiry": "2025-01-01T00:00:00Z",
  "site_name": "site1"
}' -X POST https://0.0.0.0:9001/1.0/external-site-join-token
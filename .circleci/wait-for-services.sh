#!/bin/bash
set -e

for service in nats-streaming:4222; do
    "$(dirname "$0")/wait-for-it.sh" -t 60 "$service"
done

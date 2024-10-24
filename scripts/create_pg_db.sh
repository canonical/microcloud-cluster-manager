#!/usr/bin/env bash
set -e

docker build -t pg-test -f Dockerfile.pg .

docker run --name pg-test -p 5432:5432 -d pg-test
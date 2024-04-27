#!/bin/sh
set -e

test -z "$VERSION" && {
    echo "Unable to get golangci-lint version." >&2
    exit 1
}

test -z "$GOTOOLDIR" && {
    echo "GOTOOLDIR env is empty." >&2
    exit 1
}

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOTOOLDIR" "$VERSION"

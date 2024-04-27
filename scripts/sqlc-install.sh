#!/bin/sh
set -e

PRINT_NAME="sqlc"

echo "Installing ${PRINT_NAME}"
RELEASES_URL="https://github.com/kyleconroy/sqlc/releases"
FILE_BASENAME="sqlc"

test -z "$VERSION" && {
    echo "Unable to get ${PRINT_NAME} version." >&2
    exit 1
}

test -z "$GOTOOLDIR" && {
    echo "GOTOOLDIR env is empty." >&2
    exit 1
}

ARCH=$(uname -m)
PLATFORM=$(uname -s)


test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
TAR_FILE="${TMPDIR}/${FILE_BASENAME}_${PLATFORM}_${ARCH}.tar.gz"

(
    cd "$TMPDIR"
    echo "Downloading ${PRINT_NAME} ${VERSION}..."
    SHORT_VERSION=$(echo "$VERSION" | cut -c 2-) # remove the v prefix
    curl -sfLo "$TAR_FILE" "${RELEASES_URL}/download/${VERSION}/${FILE_BASENAME}_${SHORT_VERSION}_${PLATFORM}_amd64.tar.gz"
)

echo "Extracting ${PRINT_NAME} binary..."
tar -xf "$TAR_FILE" -C "$TMPDIR"
echo "Copying binary to ${GOTOOLDIR}/${FILE_BASENAME}"
cp "${TMPDIR}/${FILE_BASENAME}" "${GOTOOLDIR}/${FILE_BASENAME}"

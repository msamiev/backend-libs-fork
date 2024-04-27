#!/bin/sh
set -e

PRINT_NAME="GoReleaser"

echo "Installing ${PRINT_NAME}"
RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
FILE_BASENAME="goreleaser"

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

# this override fixes error while script execution under docker compose on M1 arch
if [ "$ARCH" = "aarch64" ] && [ "$PLATFORM" = "Linux" ]; then
    ARCH=arm64
fi

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
TAR_FILE="${TMPDIR}/${FILE_BASENAME}_${PLATFORM}_${ARCH}.tar.gz"

(
    cd "$TMPDIR"
    echo "Downloading ${PRINT_NAME} ${VERSION}..."
    curl -sfLo "$TAR_FILE" "${RELEASES_URL}/download/${VERSION}/${FILE_BASENAME}_${PLATFORM}_${ARCH}.tar.gz"
)

echo "Extracting ${PRINT_NAME} binary..."
tar -xf "$TAR_FILE" -C "$TMPDIR"
echo "Copying binary to ${GOTOOLDIR}/${FILE_BASENAME}"
cp "${TMPDIR}/${FILE_BASENAME}" "${GOTOOLDIR}/${FILE_BASENAME}"

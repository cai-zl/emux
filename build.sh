#!/bin/sh
# Build size-optimized emux binaries for Linux and macOS (amd64 + arm64) into dist/.
set -eu
cd "$(dirname "$0")"

VERSION=$(git describe --tags --always 2>/dev/null || echo dev)
OUT=dist
LDFLAGS="-s -w -X emux/cmd.version=$VERSION"

rm -rf "$OUT"
mkdir -p "$OUT"
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
    os=${target%/*} arch=${target#*/}
    echo "building emux-$os-$arch ..."
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
        go build -trimpath -buildvcs=false -ldflags "$LDFLAGS" \
        -o "$OUT/emux-$os-$arch"
done
ls -lh "$OUT"

#!/bin/sh
# Build binary td cho Mac, Linux và Windows.
#
# Cách chạy riêng:  ./scripts/build_binaries.sh
# Thường được gọi từ build_all.sh ở thư mục gốc.
set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
. "$(dirname "$0")/td_version.sh"

APP_NAME="td"
OUT_DIR="$ROOT_DIR/out"

VERSION=$(td_get_version)
LDFLAGS=$(td_get_ldflags "$VERSION")

echo "Đang build version: $VERSION"

# Build cho các nền tảng

echo "Building for Mac (Apple Silicon)..."
cd "$ROOT_DIR" && GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "$OUT_DIR/${APP_NAME}-mac-arm-${VERSION}" .

echo "Building for Mac (Intel)..."
cd "$ROOT_DIR" && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUT_DIR/${APP_NAME}-mac-intel-${VERSION}" .

echo "Building for Linux..."
cd "$ROOT_DIR" && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUT_DIR/${APP_NAME}-linux-${VERSION}" .

echo "Building for Windows..."
cd "$ROOT_DIR" && GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$OUT_DIR/${APP_NAME}-windows-${VERSION}.exe" .

echo "Build thành công!"
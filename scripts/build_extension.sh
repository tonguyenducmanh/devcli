#!/bin/sh
# Build tiện ích VS Code thành một tệp .vsix để cài vào bất kỳ máy nào.
#
# Cách chạy:
#   ./scripts/build_extension.sh            build rồi đóng gói vào out/
#   ./scripts/build_extension.sh --out-dir thư-mục-khác
#   ./scripts/build_extension.sh --package  chỉ đóng gói, không biên dịch lại
#
# Tệp .vsix không phụ thuộc nền tảng: nó chỉ chứa mã JavaScript đã biên dịch và
# một tệp biểu tượng. Máy nào cài đều chạy được, khác ở chỗ máy đó phải có lệnh
# `tm` riêng. Xem README trong editors/vscode để biết cách đặt tm.path.
#
# Script bỏ qua nhẹ nhàng khi máy không có Node: tm là công cụ Go, việc build
# binary không được phụ thuộc vào Node. Khi đó in ra cách cài và thoát với mã 0
# để build_all.sh vẫn build được phần còn lại.
set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
EXT_DIR="$ROOT_DIR/editors/vscode"
OUT_DIR="$ROOT_DIR/out"
SKIP_BUILD=0

# ═══ Cấu hình ═══════════════════════════════════════════════════════════

# Số phiên bản của tiện ích. Nằm trong package.json của chính tiện ích, dòng
# dưới đây chỉ để đặt tên tệp cho dễ tìm.
EXT_VERSION=$(sed -n 's/.*"version": "\([^"]*\)".*/\1/p' "$EXT_DIR/package.json" | head -1)

# Tiền tố tên tệp trong out/. Đổi APP_NAME trong build_binaries.sh thì đổi cả
# ở đây, hai tệp nằm cùng nền tảng build.
APP_NAME=td-devcli

# Lệnh đóng gói. Ưu tiên bản cài sẵn trong dự án, không có thì gọi npx.
VSCE=""

# ═══ Hết cấu hình ═══════════════════════════════════════════════════════

while [ $# -gt 0 ]; do
    case "$1" in
        --package) SKIP_BUILD=1 ;;
        --out-dir) OUT_DIR="$2"; shift ;;
        *) echo "Lỗi: không hiểu tham số '$1'. Xem $0 --help" >&2; exit 1 ;;
    esac
    shift
done

skip() {
    echo "--- Bỏ qua tiện ích VS Code: $1"
    exit 0
}

command -v npm >/dev/null 2>&1 || skip "máy chưa có Node.js, xem https://nodejs.org"
command -v node >/dev/null 2>&1 || skip "máy chưa có Node.js, xem https://nodejs.org"

# Ưu tiên bản vsce cài trong thư mục tiện ích để kết quả build lặp lại được
# đúng nhau, không phụ thuộc phiên bản mới nhất trên máy.
if [ -x "$EXT_DIR/node_modules/.bin/vsce" ]; then
    VSCE="$EXT_DIR/node_modules/.bin/vsce"
fi

mkdir -p "$OUT_DIR"
OUT_FILE="$OUT_DIR/${APP_NAME}-vscode-${EXT_VERSION}.vsix"

echo "--- Build tiện ích VS Code $EXT_VERSION ---"
cd "$EXT_DIR"

if [ ! -d node_modules ]; then
    echo "Cài phụ thuộc..."
    if [ -f package-lock.json ]; then
        npm ci
    else
        npm install
    fi
fi

if [ "$SKIP_BUILD" -eq 0 ]; then
    npm run compile
fi

if [ -z "$VSCE" ]; then
    # Chưa cài vsce trong dự án thì gọi qua npx. Không cài cứng vào package.json
    # vì đó là công cụ đóng gói, không phải thứ mã tiện ích cần lúc chạy.
    echo "Dùng npx để đóng gói..."
    npx --yes @vscode/vsce package \
        --no-dependencies \
        --allow-missing-repository \
        -o "$OUT_FILE"
else
    "$VSCE" package --no-dependencies --allow-missing-repository -o "$OUT_FILE"
fi

echo "Xong: $OUT_FILE"
echo "Cài vào VS Code bằng: code --install-extension \"$OUT_FILE\""
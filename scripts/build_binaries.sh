#!/bin/sh
# Build tệp thực thi td cho các nền tảng khai báo trong TARGETS bên dưới.
#
# Cách chạy:
#   ./scripts/build_binaries.sh                  build mọi nền tảng
#   ./scripts/build_binaries.sh mac-arm          chỉ build một nền tảng
#   ./scripts/build_binaries.sh mac-arm linux    build một vài nền tảng
#   ./scripts/build_binaries.sh --list           xem danh sách nền tảng
set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

# ═══ Cấu hình ══════════════════════════════════════════════════════
#
# Toàn bộ cấu hình build nằm trong phần này. Muốn đổi cách build thì sửa ở
# đây, không đụng vào phần logic bên dưới.

# ─── Biến toàn cục của ứng dụng ────────────────────────────────────
#
# Những giá trị này được gắn vào tệp thực thi lúc biên dịch bằng cờ -X.

# Tên gọi lệnh trên terminal. Đổi thành ví dụ "devtool" thì chạy bằng
# "devtool vcs status". Tên này cũng quyết định dòng Use trong phần trợ giúp.
CMD_NAME=td

# Nơi phát hành, dùng để in gợi ý khi báo lỗi. Để rỗng thì không in gợi ý.
REPO_URL=github.com/tonguyenducmanh/devcli

# ─── Thông tin build ───────────────────────────────────────────────

# Số phiên bản của ứng dụng. Đổi ở đây mỗi khi phát hành bản mới.
# Giá trị này xuất hiện trong tên tệp trong out/ và trong kết quả `td version`.
VERSION=0.1.0

# Tên tệp thực thi, dùng làm tiền tố cho tên file trong out/.
APP_NAME=td

# Thư mục chứa kết quả build, tính từ thư mục gốc project.
OUT_DIR=out

# ─── Nền tảng cần build ────────────────────────────────────────────
#
# Định dạng mỗi dòng, phân tách bằng dấu | :
#
#     tên | goos | goarch | đuôi tệp
#
#   tên     tên ngắn gọn, dùng để chọn nền tảng khi chạy lệnh build,
#           ví dụ: ./scripts/build_all.sh mac-arm
#   goos    giá trị GOOS mà trình biên dịch Go hiểu.
#   goarch  giá trị GOARCH mà trình biên dịch Go hiểu.
#   đuôi    phần mở rộng tên tệp, bỏ trống nếu không cần.

TARGETS="
mac-arm|darwin|arm64|
linux|linux|amd64|
windows|windows|amd64|.exe
"

# ─── Cờ biên dịch ──────────────────────────────────────────────────
#
# -s -w            bỏ ký hiệu gỡ lỗi và bảng debug, tệp nhỏ hơn nhiều.
# CGO_ENABLED=0    tệp thực thi tĩnh thật sự, chạy được trên mọi máy cùng
#                  kiến trúc mà không cần thư viện hệ thống.
#
# Cờ gắn phiên bản và các biến toàn cục (-X ...cmd.Version) do
# make_ldflags bên dưới tự thêm vào, không khai báo ở đây.

BUILD_FLAGS="-s -w"
CGO_ENABLED=0

# ═══ Hết cấu hình ═════════════════════════════════════════════════

# make_ldflags tạo chuỗi ldflags gắn các biến toàn cục vào tệp thực thi.
#
# Lưu ý: phía Go phải khai báo bằng `var` thì -X mới ghi được.
make_ldflags() {
    pkg="github.com/tonguyenducmanh/devcli/cmd"
    echo "$BUILD_FLAGS"
    echo "-X ${pkg}.AppName=$CMD_NAME"
    echo "-X ${pkg}.Version=$VERSION"
    echo "-X ${pkg}.RepoURL=$REPO_URL"
}

# parse_targets đọc danh sách TARGETS và in ra từng dòng
# ở dạng: tên goos goarch đuôi
parse_targets() {
    # Tách theo | rồi nén khoảng trắng thừa, mỗi nền tảng thành một dòng.
    # Bỏ dấu gạch chéo ngược phòng khi ai đó viết dấu chấm dạng \.exe.
    echo "$TARGETS" | tr '|' ' ' | tr -s ' ' | tr -d '\\' | sed 's/^ //'
}

# build_target build một nền tảng.
# $1: tên, $2: goos, $3: goarch, $4: đuôi tệp
build_target() {
    echo "Building $1 ($2/$3)..."
    cd "$ROOT_DIR" || exit 1
    CGO_ENABLED="$CGO_ENABLED" GOOS="$2" GOARCH="$3" \
        go build -ldflags "$LDFLAGS" \
        -o "$ROOT_DIR/$OUT_DIR/${APP_NAME}-$1-${VERSION}${4}" .
}

LDFLAGS=$(make_ldflags)

mkdir -p "$ROOT_DIR/$OUT_DIR"

if [ "$1" = "--list" ]; then
    echo "Các nền tảng có thể build (phiên bản $VERSION):"
    parse_targets | while read -r name goos goarch ext; do
        [ -n "$name" ] || continue
        echo "  $name  ->  ${APP_NAME}-${name}-${VERSION}${ext}"
    done
    exit 0
fi

echo "Đang build $CMD_NAME phiên bản $VERSION"

if [ $# -eq 0 ]; then
    parse_targets | while read -r name goos goarch ext; do
        [ -n "$name" ] || continue
        build_target "$name" "$goos" "$goarch" "$ext"
    done
else
    # Build đúng các nền tảng nêu tên trên dòng lệnh.
    for want in "$@"; do
        line=$(parse_targets | grep "^$want " || true)
        if [ -z "$line" ]; then
            echo "Lỗi: không có nền tảng tên '$want'. Xem danh sách: $0 --list" >&2
            exit 1
        fi
        build_target $line
    done
fi

echo "Build thành công! Kết quả nằm trong $OUT_DIR"

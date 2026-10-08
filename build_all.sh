#!/bin/sh
# Build tm: sinh tài liệu cho trợ lý lập trình, build tệp thực thi cho mọi nền
# tảng khai báo trong scripts/build_binaries.sh, rồi đóng gói tiện ích VS Code.
#
# Script này nằm ở thư mục gốc của kho làm điểm vào duy nhất. Cấu hình build và
# các script còn lại vẫn nằm trong scripts/.
#
# Cách chạy:
#   ./build_all.sh                      sinh tài liệu, build mọi nền tảng, đóng gói tiện ích
#   ./build_all.sh mac-arm linux        chỉ build một vài nền tảng, vẫn đóng gói tiện ích
#   ./build_all.sh --no-extension       bỏ qua bước đóng gói tiện ích VS Code
#   VERSION=1.2.3 ./build_all.sh       thử phiên bản khác mà không sửa tệp
#   CMD_NAME=devtool ./build_all.sh     build với tên lệnh khác
#
# Toàn bộ cấu hình build, gồm cả tên lệnh, số phiên bản, danh sách nền tảng
# và các cờ biên dịch, nằm ở đầu scripts/build_binaries.sh.
#
# Bước 1 sinh lại agents/cli/ cho khớp với cây lệnh hiện tại. Muốn build
# tệp thực thi mà không sinh tài liệu thì gọi scripts/build_binaries.sh trực tiếp.
#
# Bước 3 đóng gói tiện ích VS Code thành một tệp .vsix đa nền tảng. Bước này cần
# Node.js; máy không có thì nó tự bỏ qua chứ không làm hỏng phần build Go, vì tm
# là công cụ Go và không được phụ thuộc vào Node.
#
# Kết quả trong out/ có chứa số phiên bản, ví dụ:
#   out/td-devcli-mac-arm-0.1.0
#   out/td-devcli-linux-0.1.0
#   out/td-devcli-windows-0.1.0.exe
#   out/td-devcli-vscode-0.1.0.vsix
set -e

# Script đang ở gốc kho nên thư mục chứa nó chính là thư mục gốc.
ROOT_DIR=$(cd "$(dirname "$0")" && pwd)

# Tách cờ của chính script này khỏi tên nền tảng trước khi chuyển tiếp, vì
# build_binaries.sh không hiểu cờ lạ nào.
BUILD_WITH_EXTENSION=1
TARGETS=""
for arg in "$@"; do
    case "$arg" in
        --no-extension) BUILD_WITH_EXTENSION=0 ;;
        *) TARGETS="$TARGETS $arg" ;;
    esac
done

"$ROOT_DIR/scripts/build_agent_docs.sh"

# # shellcheck disable=SC2086 # danh sách tên nền tảng cần tách theo khoảng trắng.
"$ROOT_DIR/scripts/build_binaries.sh" $TARGETS

if [ "$BUILD_WITH_EXTENSION" -eq 1 ]; then
    "$ROOT_DIR/scripts/build_extension.sh"
fi
#!/bin/sh
# Build td: sinh tài liệu cho trợ lý lập trình rồi build tệp thực thi cho mọi
# nền tảng khai báo trong scripts/build_binaries.sh.
#
# Script này nằm ở thư mục gốc của kho làm điểm vào duy nhất. Cấu hình build và
# các script còn lại vẫn nằm trong scripts/.
#
# Cách chạy:
#   ./build_all.sh                      sinh tài liệu rồi build mọi nền tảng
#   ./build_all.sh mac-arm linux        chỉ build một vài nền tảng
#   VERSION=1.2.3 ./build_all.sh       thử phiên bản khác mà không sửa tệp
#   CMD_NAME=devtool ./build_all.sh     build với tên lệnh khác
#
# Toàn bộ cấu hình build, gồm cả tên lệnh, số phiên bản, danh sách nền tảng
# và các cờ biên dịch, nằm ở đầu scripts/build_binaries.sh.
#
# Bước 1 sinh lại agents/cli/ cho khớp với cây lệnh hiện tại. Muốn build
# tệp thực thi mà không sinh tài liệu thì gọi scripts/build_binaries.sh trực tiếp.
#
# Kết quả trong out/ có chứa số phiên bản, ví dụ:
#   out/td-mac-arm-0.1.0
#   out/td-linux-0.1.0
#   out/td-windows-0.1.0.exe
set -e

# Script đang ở gốc kho nên thư mục chứa nó chính là thư mục gốc.
ROOT_DIR=$(cd "$(dirname "$0")" && pwd)

"$ROOT_DIR/scripts/build_agent_docs.sh"
"$ROOT_DIR/scripts/build_binaries.sh" "$@"

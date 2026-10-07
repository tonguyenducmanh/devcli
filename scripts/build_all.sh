#!/bin/sh
# Build td: sinh tài liệu cho trợ lý lập trình rồi build tệp thực thi cho mọi
# nền tảng khai báo trong scripts/build_binaries.sh.
#
# Đây là điểm vào chính để build. Chạy được từ bất kỳ thư mục nào:
#   ./scripts/build_all.sh
#   ./scripts/build_all.sh mac-arm linux      # chỉ build một vài nền tảng
#
# Toàn bộ cấu hình build, gồm cả tên lệnh, số phiên bản, danh sách nền tảng
# và các cờ biên dịch, nằm ở đầu scripts/build_binaries.sh.
#
# Bước 1 sinh lại agents/cli/ cho khớp với cây lệnh hiện tại. Muốn build
# tệp thực thi mà không sinh tài liệu thì gọi build_binaries.sh trực tiếp.
#
# Kết quả trong out/ có chứa số phiên bản, ví dụ:
#   out/td-mac-arm-0.1.0
#   out/td-linux-0.1.0
#   out/td-windows-0.1.0.exe
set -e

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

"$ROOT_DIR/scripts/build_agent_docs.sh"
"$ROOT_DIR/scripts/build_binaries.sh" "$@"

#!/bin/sh
# Build binary td cho tất cả nền tảng, kết quả nằm trong out/.
#
# Đây là điểm vào chính để build. Chạy từ thư mục gốc của project:
#   ./scripts/build_all.sh
#   VERSION=1.2.3 ./scripts/build_all.sh
#
# Phiên bản lấy từ file scripts/VERSION, hoặc ghi đè bằng biến VERSION.
#
# Kết quả trong out/ có chứa số phiên bản, ví dụ:
#   out/td-mac-arm-0.1.0
#   out/td-mac-intel-0.1.0
#   out/td-linux-0.1.0
#   out/td-windows-0.1.0.exe
set -e

# Script nằm trong scripts/ nên thư mục gốc project là thư mục cha của nó.
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
. "$ROOT_DIR/scripts/td_version.sh"

# cần build tài liệu cho trợ lý lập trình thì mở ra
# "$ROOT_DIR/scripts/build_agent_docs.sh"
"$ROOT_DIR/scripts/build_binaries.sh"
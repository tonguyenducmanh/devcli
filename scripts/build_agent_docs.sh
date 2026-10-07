#!/bin/sh
# Sinh lại toàn bộ tài liệu dành cho trợ lý lập trình vào thư mục agents/cli.
#
# Tài liệu được sinh từ cây lệnh: mỗi lệnh một tệp Markdown, gồm phần mô tả,
# cú pháp, các ví dụ và các cờ. Không sửa tay các tệp trong agents/cli.
#
# Khi nào cần chạy lại:
#   - Thêm, bỏ hoặc đổi tên một lệnh
#   - Sửa Short, Long, Example hoặc mô tả cờ của một lệnh
#   - Đổi kiến trúc cây lệnh (thêm nhóm, thêm cờ toàn cục)
#
# build_all.sh đã gọi sẵn script này nên thường không cần chạy riêng.
# Nếu quên chạy, tài liệu sẽ lệch với cây lệnh. Chạy ./scripts/check.sh để
# phát hiện ra điều đó.
set -e

# Script nằm trong scripts/ nên thư mục gốc project là thư mục cha của nó.
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
OUT_DIR="$ROOT_DIR/agents/cli"

echo "--- Sinh tài liệu lệnh vào agents/cli ---"
cd "$ROOT_DIR" && go run ./internal/tools/docgen -out "$OUT_DIR"

echo "Xong. Số tệp: $(ls -1 "$OUT_DIR" | wc -l | tr -d ' ')"

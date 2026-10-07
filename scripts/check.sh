#!/bin/sh
# Kiểm tra mã nguồn trước khi đóng góp.
#
# Chạy định dạng, phân tích tĩnh, kiểm thử và đối chiếu tài liệu với câu lệnh.
# Tất cả phải qua thì mới coi là ổn.
set -e

# Script nằm trong scripts/ nên thư mục gốc project là thư mục cha của nó.
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

echo "=== 1/4 Định dạng mã nguồn ==="
UNFORMATTED=$(gofmt -s -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "Lỗi: các tệp sau chưa được định dạng:"
    echo "$UNFORMATTED"
    echo "Sửa bằng: gofmt -s -w ."
    exit 1
fi
echo "OK"

echo "=== 2/4 Phân tích tĩnh ==="
go vet ./...
echo "OK"

echo "=== 3/4 Kiểm thử ==="
go test ./...
echo "OK"

echo "=== 4/4 Đối chiếu tài liệu ==="
# Sinh lại tài liệu vào thư mục tạm rồi so sánh với bản đã commit.
TMP_DOCS=$(mktemp -d)
trap 'rm -rf "$TMP_DOCS"' EXIT
go run ./internal/tools/docgen -out "$TMP_DOCS" >/dev/null

# So từng tệp, bỏ qua dòng thời gian sinh tự động ở cuối tệp.
if diff -r -I '^## Auto generated' "$ROOT_DIR/docs/agents/cli" "$TMP_DOCS" >/dev/null 2>&1; then
    echo "OK"
else
    echo "Lỗi: docs/agents/cli lệch với cây lệnh."
    echo "Chạy: ./scripts/build_agent_docs.sh"
    diff -r -I '^## Auto generated' "$ROOT_DIR/docs/agents/cli" "$TMP_DOCS" | head -20
    exit 1
fi

echo ""
echo "Mọi kiểm tra đều qua."
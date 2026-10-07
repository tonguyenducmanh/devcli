#!/bin/sh
# Đọc phiên bản dùng chung cho toàn bộ quá trình build của td.
#
# Nguồn version (theo thứ tự ưu tiên):
#   1. Biến môi trường VERSION khi build:  VERSION=1.2.3 ./scripts/build_all.sh
#   2. File scripts/VERSION
#
# Cách dùng:
#   . "$(dirname "$0")/td_version.sh"
#   VERSION=$(td_get_version)
#   LDFLAGS=$(td_get_ldflags "$VERSION")

# td_find_scripts_dir trả về thư mục scripts, tức là thư mục chứa file này.
# Dùng để tìm file VERSION kể cả khi script được gọi từ thư mục khác.
td_find_scripts_dir() {
    echo "$(cd "$(dirname "$0")" && pwd)"
}

# td_get_version trả về phiên bản hiện tại của app.
td_get_version() {
    if [ -n "$VERSION" ]; then
        echo "$VERSION"
        return 0
    fi

    td_scripts=$(td_find_scripts_dir)
    if [ ! -f "$td_scripts/VERSION" ]; then
        echo "Lỗi: không tìm thấy file $td_scripts/VERSION" >&2
        return 1
    fi

    # Bỏ khoảng trắng và dấu xuống dòng ở cuối
    td_version=$(tr -d ' \t\n\r' < "$td_scripts/VERSION")
    if [ -z "$td_version" ]; then
        echo "Lỗi: file VERSION rỗng" >&2
        return 1
    fi

    echo "$td_version"
}

# td_get_ldflags tạo chuỗi ldflags để gắn phiên bản vào binary.
# Biến Version trong cmd/root.go phải khai báo bằng var thì -X mới ghi được.
#   $1: phiên bản cần gắn
td_get_ldflags() {
    echo "-s -w -X github.com/tonguyenducmanh/devcli/cmd.Version=$1"
}
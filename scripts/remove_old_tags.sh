#!/bin/bash

# 1. Danh sách các tag muốn GIỮ LẠI (ngăn cách bằng khoảng trắng)
EXCLUDE_TAGS=("v16.1.5")

# Chuyển mảng thành định dạng regex để so khớp (ví dụ: ^(v1.0|v2.0|stable)$ )
EXCLUDE_PATTERN=$(printf "|%s" "${EXCLUDE_TAGS[@]}")
EXCLUDE_PATTERN="^(${EXCLUDE_PATTERN:1})$"

echo "--- Đang lấy danh sách tag từ hệ thống ---"

# 2. Lấy tất cả tag hiện có ở local
ALL_TAGS=$(git tag)

if [ -z "$ALL_TAGS" ]; then
    echo "Không tìm thấy tag nào để xóa."
    exit 0
fi

for tag in $ALL_TAGS; do
    # Kiểm tra xem tag có nằm trong danh sách loại trừ không
    if [[ $tag =~ $EXCLUDE_PATTERN ]]; then
        echo "--> Đang giữ lại tag: $tag (Bỏ qua)"
    else
        echo "--> Đang xóa tag: $tag"

        # Xóa ở Local
        git tag -d "$tag"

        # Xóa ở Remote (origin)
        git push origin --delete "$tag"
    fi
done

echo "--- Hoàn thành! ---"

# Tiện ích hệ thống (td sys)

Nhóm `sys` cung cấp các công cụ thao tác với tệp tin và hệ thống, mô phỏng lại các tiện ích cơ bản quen thuộc trên Linux mà không yêu cầu bạn phải cài đặt bất kỳ công cụ ngoại vi nào.

## Danh sách lệnh của `td sys`

| Lệnh | Công dụng |
| --- | --- |
| `ls` | Liệt kê các tệp tin trong thư mục (có cờ `-a`, `-l`) |
| `cat` | In toàn bộ nội dung của tệp ra màn hình |
| `head` | In N dòng đầu tiên của tệp (cờ `-n`) |
| `tail` | In N dòng cuối cùng của tệp (cờ `-n`) |
| `pwd` | In đường dẫn tuyệt đối của thư mục làm việc hiện tại |
| `mkdir` | Tạo thư mục mới (cờ `-p` để tạo đệ quy các thư mục cha) |
| `touch` | Tạo tệp tin trống hoặc cập nhật thời gian sửa đổi |
| `rm` | Xoá tệp hoặc thư mục (cờ `-r` xoá đệ quy, `-f` xoá cưỡng chế) |
| `cp` | Sao chép tệp hoặc thư mục (cờ `-r` để sao chép đệ quy) |
| `mv` | Di chuyển hoặc đổi tên tệp/thư mục |
| `wc` | Đếm số dòng (`-l`), số từ (`-w`), số byte (`-c`) của tệp |
| `grep` | Tìm kiếm các chuỗi văn bản khớp biểu thức chính quy |
| `rmempty` | Tìm và xoá đệ quy tất cả các thư mục rỗng |

## Sử dụng mặc định

Bạn có thể thiết lập `sys` làm nhóm lệnh mặc định thông qua lệnh `td use`:

```bash
td use sys
```

Sau khi thiết lập, bạn có thể gọi thẳng các lệnh mà không cần qua tiền tố `td sys`, ví dụ:
```bash
td ls
td grep "hello" file.txt
td rm -rf build/
```

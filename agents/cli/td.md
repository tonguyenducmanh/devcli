## td

Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất

### Synopsis

td gom mọi công cụ dòng lệnh bạn dùng hằng ngày dưới một lệnh duy nhất.

Mỗi nhóm công cụ là một lệnh con của td, ví dụ:
  td vcs ...     quản lý phiên bản mã nguồn cục bộ

Mỗi nhóm tự quản lý toàn bộ dữ liệu của nó trong thư mục riêng cạnh dự án, nên
td không cần cài thêm hay cấu hình gì cả.

```
td [flags]
```

### Examples

```
  td vcs init                   khởi tạo kho tại thư mục hiện tại
  td vcs status                 xem các thay đổi chưa commit
  td vcs commit -m "tin nhắn"   ghi lại thay đổi
  td config --list              xem cấu hình đang dùng

  td                            xem phiên bản và danh sách lệnh
  td -v                         xem thêm thông tin môi trường
  td --help                     xem trợ giúp đầy đủ
```

### Options

```
  -h, --help       hiển thị phần trợ giúp của lệnh này
      --no-color   tắt màu trong output
  -v, --verbose    in thêm thông tin chi tiết
      --version    hiển thị phiên bản rồi thoát
```

### SEE ALSO

* [td config](td_config.md)	 - Xem và chỉnh sửa cấu hình của td
* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ
* [td version](td_version.md)	 - Hiển thị phiên bản và thông tin môi trường


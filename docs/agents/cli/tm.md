## tm

Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất

### Synopsis

tm gom mọi công cụ dòng lệnh bạn dùng hằng ngày dưới một lệnh duy nhất.

Mỗi nhóm công cụ là một lệnh con của tm, ví dụ:
  tm vcs ...     quản lý phiên bản mã nguồn cục bộ

Mỗi nhóm tự quản lý toàn bộ dữ liệu của nó trong thư mục riêng cạnh dự án, nên
tm không cần cài thêm hay cấu hình gì cả.

```
tm [flags]
```

### Examples

```
  tm vcs init                   khởi tạo kho tại thư mục hiện tại
  tm vcs status                 xem các thay đổi chưa commit
  tm vcs commit -m "tin nhắn"   ghi lại thay đổi
  tm config --list              xem cấu hình đang dùng

  tm                            xem phiên bản và danh sách lệnh
  tm -v                         xem thêm thông tin môi trường
  tm --help                     xem trợ giúp đầy đủ
```

### Options

```
  -h, --help      hiển thị phần trợ giúp của lệnh này
  -v, --verbose   in thêm thông tin môi trường
      --version   hiển thị phiên bản rồi thoát
```

### SEE ALSO

* [tm config](config/tm_config.md)	 - Xem và chỉnh sửa cấu hình của tm
* [tm sys](sys/tm_sys.md)	 - Các tiện ích hệ thống (ls, cat, head, tail)
* [tm use](use/tm_use.md)	 - Đặt nhóm công cụ mặc định
* [tm vcs](vcs/tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ
* [tm version](version/tm_version.md)	 - Hiển thị phiên bản và thông tin môi trường


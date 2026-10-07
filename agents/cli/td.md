## td

td - bộ công cụ dòng lệnh cá nhân

### Synopsis

td là bộ công cụ dòng lệnh cá nhân, tự quản lý toàn bộ dữ liệu của nó.

Mỗi nhóm công cụ là một lệnh con của td, ví dụ:
  td vcs ...     quản lý phiên bản mã nguồn cục bộ

Dữ liệu của td được lưu trong thư mục .tdx cạnh dự án.

```
td [flags]
```

### Examples

```
  td vcs init                     khởi tạo kho tại thư mục hiện tại
  td vcs status                   xem các thay đổi chưa commit
  td vcs commit -m "tin nhắn"     ghi lại thay đổi
  td config --list                xem cấu hình đang dùng
  td --help                       xem toàn bộ lệnh
```

### Options

```
  -h, --help       help for td
      --no-color   tắt màu trong output
  -v, --verbose    in thêm thông tin chi tiết
```

### SEE ALSO

* [td config](td_config.md)	 - Xem và chỉnh sửa cấu hình của td
* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ
* [td version](td_version.md)	 - Hiển thị phiên bản của td


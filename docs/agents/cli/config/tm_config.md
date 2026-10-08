## tm config

Xem và chỉnh sửa cấu hình của tm

### Synopsis

Cấu hình được lưu ở hai nơi: file toàn cục và file trong từng repo.
Cấu hình toàn cục được đọc trước, nên nên đặt user.name và user.email ở đó.

```
tm config [flags]
```

### Examples

```
  tm config --global user.name "Tên của tôi"
  tm config user.email "ten@example.com"
  tm config --list
```

### Options

```
  -C, --dir string   chạy lệnh tại thư mục khác
  -g, --global       áp dụng cho toàn bộ máy thay vì repo hiện tại
  -h, --help         hiển thị phần trợ giúp của lệnh này
  -l, --list         liệt kê toàn bộ cấu hình
      --unset        xóa một khóa cấu hình
  -v, --verbose      in ra tệp cấu hình đã dùng
```

### SEE ALSO

* [tm](../tm.md)	 - Bộ công cụ dòng lệnh cá nhân, tất cả gói dưới một lệnh duy nhất


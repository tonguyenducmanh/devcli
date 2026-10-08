## tm vcs tag

Liệt kê, tạo hoặc xóa tag

### Synopsis

Đánh dấu một điểm trong lịch sử bằng tên dễ nhớ, ví dụ theo phiên bản.

Tag nhẹ chỉ là một con trỏ trỏ tới commit. Tag có chú thích (-a kèm -m) tạo
thêm một object riêng nên lưu được lời giải thích, ai đó và thời điểm tạo.

Một tham số là tạo tag tại HEAD, không có tham số thì in danh sách.

Dùng -q khi cần đọc danh sách bằng kịch bản: lệnh sẽ in ra rỗng thay vì in
thông báo "Chưa có tag nào", nên không phải lọc bỏ câu văn bản.

```
tm vcs tag [flags]
```

### Examples

```
  tm vcs tag                        liệt kê các tag
  tm vcs tag v1.0.0                 tạo tag nhẹ
  tm vcs tag -a v1.0.0 -m "ghi chú"  tạo tag có chú thích
  tm vcs tag -d v1.0.0              xóa tag
```

### Options

```
  -a, --annotate         tạo tag có chú thích
  -d, --delete           xóa tag
  -h, --help             hiển thị phần trợ giúp của lệnh này
  -l, --list             liệt kê các tag
  -m, --message string   nội dung chú thích cho tag
  -q, --quiet            liệt kê và im lặng khi không có tag nào, để dùng trong kịch bản
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


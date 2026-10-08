## tm vcs clean

Xoá tệp chưa được theo dõi

### Synopsis

Xoá khỏi cây làm việc những tệp tm chưa từng theo dõi.

Tệp đã được đưa vào vùng chuẩn bị thì thuộc về lịch sử nên lệnh này không đụng
tới, muốn bỏ theo dõi thì dùng `tm vcs rm`. Nhờ vậy lệnh không bao giờ
xoá mất thứ còn cứu được trong kho.

Đối số là tệp, thư mục hoặc mẫu có dấu * và ?, khớp với cách `tm vcs add`
hiểu đường dẫn. Không có đối số thì áp dụng cho mọi tệp chưa theo dõi.

Lệnh không hỏi lại: không có cờ -f thì chỉ liệt kê những gì sẽ bị xoá, đọc xong
xác nhận trong danh sách rồi chạy lại với -f mới xoá thật.

```
tm vcs clean [tệp...] [flags]
```

### Examples

```
  tm vcs clean              liệt kê các tệp sẽ bị xoá
  tm vcs clean -f           xoá mọi tệp chưa được theo dõi
  tm vcs clean -f build/    xoá tệp chưa theo dõi trong thư mục build
  tm vcs clean -f "*.log"   xoá mọi tệp kết thúc bằng .log
```

### Options

```
      --dry-run   chỉ liệt kê, giống khi không kèm -f
  -f, --force     xoá thật thay vì chỉ liệt kê
  -h, --help      hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


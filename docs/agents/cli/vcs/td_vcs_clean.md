## td vcs clean

Xoá tệp chưa được theo dõi

### Synopsis

Xoá khỏi cây làm việc những tệp td chưa từng theo dõi.

Tệp đã được đưa vào vùng chuẩn bị thì thuộc về lịch sử nên lệnh này không đụng
tới, muốn bỏ theo dõi thì dùng `td vcs rm`. Nhờ vậy lệnh không bao giờ
xoá mất thứ còn cứu được trong kho.

Đối số là tệp, thư mục hoặc mẫu có dấu * và ?, khớp với cách `td vcs add`
hiểu đường dẫn. Không có đối số thì áp dụng cho mọi tệp chưa theo dõi.

Lệnh không hỏi lại: không có cờ -f thì chỉ liệt kê những gì sẽ bị xoá, đọc xong
xác nhận trong danh sách rồi chạy lại với -f mới xoá thật.

```
td vcs clean [tệp...] [flags]
```

### Examples

```
  td vcs clean              liệt kê các tệp sẽ bị xoá
  td vcs clean -f           xoá mọi tệp chưa được theo dõi
  td vcs clean -f build/    xoá tệp chưa theo dõi trong thư mục build
  td vcs clean -f "*.log"   xoá mọi tệp kết thúc bằng .log
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

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


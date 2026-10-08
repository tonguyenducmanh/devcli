## tm vcs log

Xem lịch sử commit

### Synopsis

Duyệt lịch sử commit từ một điểm bắt đầu đi ngược về các commit cha.

Mặc định bắt đầu từ HEAD. Sau dấu hai gạch ngang là danh sách tệp, khi đó chỉ
những commit có thay đổi tệp đó mới được hiển thị.

Cột đầu là mã băm ngắn, kèm các tham chiếu đang trỏ tới commit đó; HEAD được
đánh dấu bằng HEAD -> để phân biệt với các nhánh khác.

```
tm vcs log [flags]
```

### Examples

```
  tm vcs log
  tm vcs log --oneline
  tm vcs log -n 5 --patch
  tm vcs log --all
  tm vcs log -- cmd/            chỉ xem các commit có sửa thư mục cmd/
```

### Options

```
  -a, --all             duyệt lịch sử của mọi nhánh và tag
  -h, --help            hiển thị phần trợ giúp của lệnh này
  -n, --max-count int   giới hạn số commit hiển thị
  -l, --oneline         mỗi commit trên một dòng
  -p, --patch           kèm nội dung diff của từng commit
      --reverse         in ngược thứ tự thời gian
      --stat            kèm thống kê số dòng thay đổi
  -v, --verbose         in cả nội dung thay đổi của từng commit
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


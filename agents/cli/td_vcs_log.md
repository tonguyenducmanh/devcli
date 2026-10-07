## td vcs log

Xem lịch sử commit

### Synopsis

Duyệt lịch sử commit từ một điểm bắt đầu đi ngược về các commit cha.

Mặc định bắt đầu từ HEAD. Sau dấu hai gạch ngang là danh sách tệp, khi đó chỉ
những commit có thay đổi tệp đó mới được hiển thị.

Cột đầu là mã băm ngắn, kèm các tham chiếu đang trỏ tới commit đó; HEAD được
đánh dấu bằng HEAD -> để phân biệt với các nhánh khác.

```
td vcs log [flags]
```

### Examples

```
  td vcs log
  td vcs log --oneline
  td vcs log -n 5 --patch
  td vcs log --all
  td vcs log -- cmd/            chỉ xem các commit có sửa thư mục cmd/
```

### Options

```
  -a, --all             duyệt lịch sử của mọi nhánh và tag
  -h, --help            help for log
  -n, --max-count int   giới hạn số commit hiển thị
  -l, --oneline         mỗi commit trên một dòng
  -p, --patch           kèm nội dung diff của từng commit
      --reverse         in ngược thứ tự thời gian
      --stat            kèm thống kê số dòng thay đổi
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


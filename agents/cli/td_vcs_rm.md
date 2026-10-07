## td vcs rm

Gỡ tệp khỏi theo dõi và khỏi cây làm việc

### Synopsis

Xoá tệp khỏi vùng chuẩn bị và khỏi cây làm việc.

Thay đổi được ghi ở lần commit kế tiếp, tệp vẫn còn trong lịch sử. Kèm
--cached để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên tệp trên đĩa.

Tệp đã xoá có thể đưa lại bằng lệnh restore trước khi commit.

```
td vcs rm <tệp>... [flags]
```

### Examples

```
  td vcs rm tệp-cũ.txt
  td vcs rm thư-mục/tệp.txt
  td vcs rm --cached tệp-vẫn-giữ.txt
```

### Options

```
      --cached   chỉ gỡ khỏi theo dõi, giữ file trên đĩa
  -h, --help     help for rm
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


## tm vcs rm

Gỡ tệp khỏi theo dõi và khỏi cây làm việc

### Synopsis

Xoá tệp khỏi vùng chuẩn bị và khỏi cây làm việc.

Thay đổi được ghi ở lần commit kế tiếp, tệp vẫn còn trong lịch sử. Kèm
--cached để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên tệp trên đĩa.

Tệp đã xoá có thể đưa lại bằng lệnh restore trước khi commit.

```
tm vcs rm <tệp>... [flags]
```

### Examples

```
  tm vcs rm tệp-cũ.txt
  tm vcs rm thư-mục/tệp.txt
  tm vcs rm --cached tệp-vẫn-giữ.txt
```

### Options

```
      --cached   chỉ gỡ khỏi theo dõi, giữ file trên đĩa
  -h, --help     hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


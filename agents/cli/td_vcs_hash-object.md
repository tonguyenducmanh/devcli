## td vcs hash-object

Tính và in mã băm của nội dung tệp

### Synopsis

Đọc nội dung tệp trên đĩa rồi in mã băm tương ứng.

Lệnh không ghi gì vào kho, chỉ cho biết mã băm mà td sẽ dùng nếu tệp đó được
đưa vào kho. Hai tệp có cùng nội dung sẽ cho cùng một mã băm.

```
td vcs hash-object <tệp>... [flags]
```

### Examples

```
  td vcs hash-object main.go
  td vcs hash-object tệp-một tệp-hai
```

### Options

```
  -h, --help   help for hash-object
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


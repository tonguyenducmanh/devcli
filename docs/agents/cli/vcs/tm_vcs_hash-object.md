## tm vcs hash-object

Tính và in mã băm của nội dung tệp

### Synopsis

Đọc nội dung tệp trên đĩa rồi in mã băm tương ứng.

Lệnh không ghi gì vào kho, chỉ cho biết mã băm mà tm sẽ dùng nếu tệp đó được
đưa vào kho. Hai tệp có cùng nội dung sẽ cho cùng một mã băm.

```
tm vcs hash-object <tệp>... [flags]
```

### Examples

```
  tm vcs hash-object main.go
  tm vcs hash-object tệp-một tệp-hai
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


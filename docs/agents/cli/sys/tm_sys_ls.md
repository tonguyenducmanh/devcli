## tm sys ls

Liệt kê các tệp tin trong thư mục

### Synopsis

Liệt kê các tệp tin và thư mục con trong thư mục được chỉ định.
Nếu không truyền thư mục, mặc định sẽ liệt kê thư mục hiện tại.

```
tm sys ls [thư mục] [flags]
```

### Examples

```
  tm sys ls
  tm sys ls -l
  tm sys ls -a /tmp
```

### Options

```
  -a, --all    hiển thị cả tệp ẩn
  -h, --help   hiển thị phần trợ giúp của lệnh này
  -l, --long   hiển thị chi tiết (quyền, kích thước, ngày giờ)
```

### SEE ALSO

* [tm sys](tm_sys.md)	 - Các tiện ích hệ thống (ls, cat, head, tail)


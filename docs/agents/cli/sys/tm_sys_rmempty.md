## tm sys rmempty

Xoá các thư mục rỗng đệ quy

### Synopsis

Tìm và xoá tất cả các thư mục rỗng bên trong thư mục được chỉ định.
Quá trình này được thực hiện đệ quy (xoá thư mục con rỗng, sau đó nếu thư mục cha rỗng thì xoá tiếp).
Mặc định sẽ quét thư mục hiện tại.

```
tm sys rmempty [thư mục] [flags]
```

### Examples

```
  tm sys rmempty
  tm sys rmempty /tmp/test
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### SEE ALSO

* [tm sys](tm_sys.md)	 - Các tiện ích hệ thống (ls, cat, head, tail)


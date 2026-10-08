## tm sys head

In N dòng đầu tiên của tệp

### Synopsis

Đọc và in ra N dòng đầu tiên của một tệp văn bản.
Mặc định in 10 dòng đầu tiên.

```
tm sys head [tệp] [flags]
```

### Examples

```
  tm sys head file.txt
  tm sys head -n 5 file.txt
```

### Options

```
  -h, --help        hiển thị phần trợ giúp của lệnh này
  -n, --lines int   số dòng cần in (default 10)
```

### SEE ALSO

* [tm sys](tm_sys.md)	 - Các tiện ích hệ thống (ls, cat, head, tail)


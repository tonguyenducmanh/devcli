## td sys tail

In N dòng cuối cùng của tệp

### Synopsis

Đọc và in ra N dòng cuối cùng của một tệp văn bản.
Mặc định in 10 dòng cuối cùng.

```
td sys tail [tệp] [flags]
```

### Examples

```
  td sys tail file.txt
  td sys tail -n 5 file.txt
```

### Options

```
  -h, --help        hiển thị phần trợ giúp của lệnh này
  -n, --lines int   số dòng cần in (default 10)
```

### SEE ALSO

* [td sys](td_sys.md)	 - Các tiện ích hệ thống (ls, cat, head, tail)


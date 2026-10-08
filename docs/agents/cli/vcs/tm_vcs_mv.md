## tm vcs mv

Đổi tên hoặc di chuyển một tệp đã được theo dõi

### Synopsis

Di chuyển hoặc đổi tên một tệp trong cây làm việc và cập nhật vùng chuẩn
bị theo đường dẫn mới.

Thư mục đích không cần tồn tại trước, được tạo tự động.

```
tm vcs mv <nguồn> <đích> [flags]
```

### Examples

```
  tm vcs mv tên-cũ.txt tên-mới.txt
  tm vcs mv tệp.txt thư-mục/tệp.txt
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


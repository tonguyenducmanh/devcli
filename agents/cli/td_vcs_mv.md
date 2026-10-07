## td vcs mv

Đổi tên hoặc di chuyển một tệp đã được theo dõi

### Synopsis

Di chuyển hoặc đổi tên một tệp trong cây làm việc và cập nhật vùng chuẩn
bị theo đường dẫn mới.

Thư mục đích không cần tồn tại trước, được tạo tự động.

```
td vcs mv <nguồn> <đích> [flags]
```

### Examples

```
  td vcs mv tên-cũ.txt tên-mới.txt
  td vcs mv tệp.txt thư-mục/tệp.txt
```

### Options

```
  -h, --help   help for mv
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


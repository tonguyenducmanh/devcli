## tm vcs switch

Chuyển sang nhánh khác

### Synopsis

Chuyển nhánh đang làm việc. Lệnh rút gọn của checkout dành cho trường hợp
chỉ cần đổi nhánh, không cần thêm tuỳ chọn nào khác.

```
tm vcs switch <nhánh> [flags]
```

### Examples

```
  tm vcs switch main
  tm vcs switch -c moi        tạo nhánh moi rồi chuyển sang đó
```

### Options

```
  -c, --create string     tạo nhánh mới rồi chuyển sang đó
      --discard-changes   bỏ qua các thay đổi chưa lưu
  -h, --help              hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


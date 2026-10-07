## td vcs switch

Chuyển sang nhánh khác

### Synopsis

Chuyển nhánh đang làm việc. Lệnh rút gọn của checkout dành cho trường hợp
chỉ cần đổi nhánh, không cần thêm tuỳ chọn nào khác.

```
td vcs switch <nhánh> [flags]
```

### Examples

```
  td vcs switch main
  td vcs switch -c moi        tạo nhánh moi rồi chuyển sang đó
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
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


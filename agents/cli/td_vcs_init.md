## td vcs init

Khởi tạo kho mã nguồn td trong thư mục cho trước

### Synopsis

Tạo thư mục .tdx trong thư mục cho trước để bắt đầu theo dõi phiên bản.

Tham số thư mục không bắt buộc, mặc định là thư mục hiện tại. Lệnh sẽ báo lỗi
nếu thư mục đó đã có kho, để tránh ghi đè dữ liệu đang có.

```
td vcs init [thư mục] [flags]
```

### Examples

```
  # Tạo kho trong thư mục hiện tại với nhánh main
  td vcs init

  # Tạo kho trong một thư mục khác với tên nhánh khác
  td vcs init du-an-cua-toi --initial-branch=develop
```

### Options

```
  -h, --help                    help for init
      --initial-branch string   tên nhánh khởi tạo (default "main")
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


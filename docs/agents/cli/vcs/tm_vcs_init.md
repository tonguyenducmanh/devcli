## tm vcs init

Khởi tạo kho mã nguồn tm trong thư mục cho trước

### Synopsis

Tạo thư mục .tmx trong thư mục cho trước để bắt đầu theo dõi phiên bản.

Tham số thư mục không bắt buộc, mặc định là thư mục hiện tại. Lệnh sẽ báo lỗi
nếu thư mục đó đã có kho, để tránh ghi đè dữ liệu đang có.

```
tm vcs init [thư mục] [flags]
```

### Examples

```
  # Tạo kho trong thư mục hiện tại với nhánh main
  tm vcs init

  # Tạo kho trong một thư mục khác với tên nhánh khác
  tm vcs init du-an-cua-toi --initial-branch=develop
```

### Options

```
  -h, --help                    hiển thị phần trợ giúp của lệnh này
      --initial-branch string   tên nhánh khởi tạo (default "main")
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


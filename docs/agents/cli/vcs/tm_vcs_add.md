## tm vcs add

Đưa thay đổi vào vùng chuẩn bị commit

### Synopsis

Đưa nội dung hiện tại của tệp vào vùng chuẩn bị, nơi nội dung được
chốt lại cho commit kế tiếp.

Không có đối số thì chỉ cập nhật các tệp đã được theo dõi. Có thể truyền
đường dẫn cụ thể, một thư mục, hoặc mẫu có dấu * và ?.

Tệp bị xoá khỏi đĩa cũng được gỡ khỏi vùng chuẩn bị. Tệp chưa theo dõi là
tệp tm chưa quản lý, thêm vào .tmxignore nếu muốn bỏ qua vĩnh viễn.

```
tm vcs add [tệp...] [flags]
```

### Examples

```
  tm vcs add .              thêm mọi thay đổi
  tm vcs add main.go        thêm một tệp
  tm vcs add --update .     chỉ cập nhật tệp đã được theo dõi
  tm vcs add "docs/*.md"    thêm theo mẫu đường dẫn
```

### Options

```
  -A, --all       đưa mọi thay đổi vào stage
  -h, --help      hiển thị phần trợ giúp của lệnh này
  -u, --update    chỉ cập nhật tệp đã được theo dõi
  -v, --verbose   in ra từng tệp đã đưa vào vùng chuẩn bị
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


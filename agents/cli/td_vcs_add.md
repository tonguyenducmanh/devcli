## td vcs add

Đưa thay đổi vào vùng chuẩn bị commit

### Synopsis

Đưa nội dung hiện tại của tệp vào vùng chuẩn bị, nơi nội dung được
chốt lại cho commit kế tiếp.

Không có đối số thì chỉ cập nhật các tệp đã được theo dõi. Có thể truyền
đường dẫn cụ thể, một thư mục, hoặc mẫu có dấu * và ?.

Tệp bị xoá khỏi đĩa cũng được gỡ khỏi vùng chuẩn bị. Tệp chưa theo dõi là
tệp td chưa quản lý, thêm vào .tdxignore nếu muốn bỏ qua vĩnh viễn.

```
td vcs add [tệp...] [flags]
```

### Examples

```
  td vcs add .              thêm mọi thay đổi
  td vcs add main.go        thêm một tệp
  td vcs add --update .     chỉ cập nhật tệp đã được theo dõi
  td vcs add "docs/*.md"    thêm theo mẫu đường dẫn
```

### Options

```
  -A, --all      đưa mọi thay đổi vào stage
  -h, --help     help for add
  -u, --update   chỉ cập nhật tệp đã được theo dõi
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


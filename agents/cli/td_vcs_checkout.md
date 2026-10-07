## td vcs checkout

Chuyển sang nhánh hoặc commit khác

### Synopsis

Di chuyển con trỏ HEAD sang nhánh, tag hoặc commit khác, đồng thời đưa vùng
chuẩn bị và cây làm việc về đúng nội dung của đích.

Khi chuyển sang một commit cụ thể, HEAD trở nên tách rời khỏi nhánh. Các thay
đổi chưa lưu sẽ không bị ghi đè, lệnh báo lỗi để bạn xử lý trước.

```
td vcs checkout [flags]
```

### Examples

```
  td vcs checkout main           chuyển sang nhánh main
  td vcs checkout -b moi         tạo nhánh moi rồi chuyển sang đó
  td vcs checkout abc1234        chuyển tới một commit cụ thể (HEAD tách rời)
  td vcs checkout --detach main  chuyển tới commit của main
```

### Options

```
  -b, --branch string   tạo nhánh mới rồi chuyển sang đó
      --detach          chuyển tới một commit, không gắn với nhánh
      --force           ghi đè các thay đổi chưa lưu trong cây làm việc
  -h, --help            help for checkout
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


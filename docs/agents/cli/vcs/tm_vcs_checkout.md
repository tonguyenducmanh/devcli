## tm vcs checkout

Chuyển sang nhánh hoặc commit khác

### Synopsis

Di chuyển con trỏ HEAD sang nhánh, tag hoặc commit khác, đồng thời đưa vùng
chuẩn bị và cây làm việc về đúng nội dung của đích.

Khi chuyển sang một commit cụ thể, HEAD trở nên tách rời khỏi nhánh. Các thay
đổi chưa lưu sẽ không bị ghi đè, lệnh báo lỗi để bạn xử lý trước.

```
tm vcs checkout [flags]
```

### Examples

```
  tm vcs checkout main           chuyển sang nhánh main
  tm vcs checkout -b moi         tạo nhánh moi rồi chuyển sang đó
  tm vcs checkout abc1234        chuyển tới một commit cụ thể (HEAD tách rời)
  tm vcs checkout --detach main  chuyển tới commit của main
```

### Options

```
  -b, --branch string   tạo nhánh mới rồi chuyển sang đó
      --detach          chuyển tới một commit, không gắn với nhánh
      --force           ghi đè các thay đổi chưa lưu trong cây làm việc
  -h, --help            hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


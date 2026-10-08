## tm vcs revert

Hoàn tác thay đổi của một commit

### Synopsis

Hoàn tác thay đổi của một commit bằng cách tạo một commit mới.

Lịch sử không bị viết lại nên lệnh này an toàn với nhánh đã chia sẻ. Khi hoàn
tác nhiều commit, chúng được xử lý theo thứ tự ngược: commit mới nhất trước.

```
tm vcs revert <commit>... [flags]
```

### Examples

```
  tm vcs revert abc1234
  tm vcs revert --no-commit abc1234
  tm vcs revert --abort
```

### Options

```
      --abort       huỷ thao tác đang dở dang
      --continue    tiếp tục sau khi giải quyết xung đột
  -h, --help        hiển thị phần trợ giúp của lệnh này
  -n, --no-commit   chỉ áp dụng thay đổi vào vùng stage
      --skip        bỏ qua commit hiện tại
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [tm vcs](tm_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


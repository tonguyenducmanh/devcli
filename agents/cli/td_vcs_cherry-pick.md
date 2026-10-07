## td vcs cherry-pick

Áp dụng thay đổi của một commit cụ thể

### Synopsis

Chép thay đổi của một hoặc nhiều commit từ nhánh khác vào nhánh đang đứng.

Mỗi commit được áp dụng như một lần hợp nhất ba phía, nên thay đổi được giữ
nguyên dù nhánh nguồn đã tiến xa. Truyền nhiều mã băm để áp dụng theo đúng
thứ tự đã cho.

```
td vcs cherry-pick <commit>... [flags]
```

### Examples

```
  td vcs cherry-pick abc1234
  td vcs cherry-pick abc1234 def5678
  td vcs cherry-pick --no-commit abc1234   chỉ áp dụng vào vùng stage
  td vcs cherry-pick --abort                huỷ khi đang giải quyết xung đột
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

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


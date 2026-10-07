## td vcs reflog

Xem nhật ký di chuyển của HEAD hoặc một tham chiếu

### Synopsis

Mỗi lần một tham chiếu dịch chuyển sẽ được ghi lại kèm mã băm cũ,
mã băm mới và lý do. Lệnh in danh sách từ mục mới nhất trở về.

Mục cũ vẫn nằm trong kho nên có thể quay lại bằng cách trỏ một tham chiếu
tới mã băm tương ứng, ví dụ: td vcs switch abc1234

```
td vcs reflog [ref] [flags]
```

### Examples

```
  td vcs reflog
  td vcs reflog main
  td vcs reflog refs/tags/v1.0.0
```

### Options

```
  -h, --help   hiển thị phần trợ giúp của lệnh này
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ


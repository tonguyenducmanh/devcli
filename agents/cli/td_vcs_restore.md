## td vcs restore

Khôi phục lại nội dung tệp

### Synopsis

Đưa lại nội dung tệp từ một nguồn khác về cây làm việc hoặc vùng chuẩn bị.

Mặc định lấy nội dung đang có trong vùng chuẩn bị, tức là huỷ các sửa đổi chưa
stage. Dùng --staged để chỉ gỡ khỏi vùng chuẩn bị mà giữ nguyên cây làm việc,
và --source để lấy từ một commit bất kỳ.

```
td vcs restore <tệp>... [flags]
```

### Examples

```
  td vcs restore main.go         lấy lại nội dung đang stage
  td vcs restore --staged main.go  gỡ thay đổi đã stage
  td vcs restore --source=HEAD~1 main.go  lấy từ một commit khác
```

### Options

```
  -h, --help            help for restore
  -s, --source string   lấy nội dung từ một commit thay vì index
  -S, --staged          chỉ thay đổi vùng stage, giữ nguyên cây làm việc
  -W, --worktree        chỉ thay đổi cây làm việc
```

### Options inherited from parent commands

```
  -C, --dir string   chạy lệnh tại thư mục khác
      --no-color     tắt màu trong output
  -v, --verbose      in thêm thông tin chi tiết
```

### SEE ALSO

* [td vcs](td_vcs.md)	 - Quản lý phiên bản mã nguồn cục bộ

